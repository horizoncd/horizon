// forked from https://github.com/kubernetes/dashboard/blob/master/src/app/backend/handler/terminal.go ,
// change restful.Request to http.Request

package terminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"gopkg.in/igm/sockjs-go.v3/sockjs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const EndOfTransmission = "\u0004"

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

// Session implements PtyHandler (using a SockJS connection)
type Session struct {
	id            string
	bound         chan error
	sockJSSession sockjs.Session
	sizeChan      chan remotecommand.TerminalSize
	doneChan      chan struct{}
}

// Message is the messaging protocol between ShellController and TerminalSession.
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// bind    fe->be     SessionID      Id sent back from TerminalResponse
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type Message struct {
	Op, Data, SessionID string
	Rows, Cols          uint16
}

// TerminalSize handles pty->process resize events
// Called in a loop from remotecommand as long as the process is running
func (t Session) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	case <-t.doneChan:
		return nil
	}
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remotecommand as long as the process is running
func (t Session) Read(p []byte) (int, error) {
	m, err := t.sockJSSession.Recv()
	if err != nil {
		// Send terminated signal to process to avoid resource leak
		return copy(p, EndOfTransmission), err
	}

	var msg Message
	if err := json.Unmarshal([]byte(m), &msg); err != nil {
		return copy(p, EndOfTransmission), err
	}

	switch msg.Op {
	case "stdin":
		return copy(p, msg.Data), nil
	case "resize":
		t.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		return copy(p, EndOfTransmission), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (t Session) Write(p []byte) (int, error) {
	msg, err := json.Marshal(Message{
		Op:   "stdout",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}

	if err = t.sockJSSession.Send(string(msg)); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t Session) Toast(p string) error {
	msg, err := json.Marshal(Message{
		Op:   "toast",
		Data: p,
	})
	if err != nil {
		return err
	}

	if err = t.sockJSSession.Send(string(msg)); err != nil {
		return err
	}
	return nil
}

// SessionMap stores a map of all TerminalSession objects and a lock to avoid concurrent conflict
type SessionMap struct {
	Sessions map[string]Session
	Lock     sync.RWMutex
}

// Get return a given terminalSession by sessionId
func (sm *SessionMap) Get(sessionID string) Session {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	return sm.Sessions[sessionID]
}

// Set store a TerminalSession to SessionMap
func (sm *SessionMap) Set(sessionID string, session Session) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	sm.Sessions[sessionID] = session
}

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (sm *SessionMap) Close(sessionID string, status uint32, reason string) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()

	var session Session
	var exists bool
	if session, exists = sm.Sessions[sessionID]; !exists {
		return
	}

	if session.sockJSSession != nil {
		err := session.sockJSSession.Close(status, reason)
		if err != nil {
			log.Println(err)
		}
	}

	delete(sm.Sessions, sessionID)
}

var terminalSessions = SessionMap{Sessions: make(map[string]Session)}

// handleTerminalSession is Called by net/http for any new /api/sockjs connections
func handleTerminalSession(session sockjs.Session) {
	var (
		buf             string
		err             error
		msg             Message
		terminalSession Session
	)

	if buf, err = session.Recv(); err != nil {
		log.Printf("handleTerminalSession: can't Recv: %v", err)
		return
	}

	if err = json.Unmarshal([]byte(buf), &msg); err != nil {
		log.Printf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
		return
	}

	if msg.Op != "bind" {
		log.Printf("handleTerminalSession: expected 'bind' message, got: %s", buf)
		return
	}

	if terminalSession = terminalSessions.Get(msg.SessionID); terminalSession.id == "" {
		log.Printf("handleTerminalSession: can't find session '%s'", msg.SessionID)
		return
	}

	terminalSession.sockJSSession = session
	terminalSessions.Set(msg.SessionID, terminalSession)
	terminalSession.bound <- nil
}

// CreateAttachHandler is called from main for /api/sockjs
func CreateAttachHandler(path string) http.Handler {
	return sockjs.NewHandler(path, sockjs.DefaultOptions, handleTerminalSession)
}

// startProcess is called by handleAttach
// Executed cmd in the container specified in request and
// connects it up with the ptyHandler (a session)
func startProcess(k8sClient kubernetes.Interface, cfg *rest.Config,
	ref ContainerRef, cmd []string, ptyHandler PtyHandler) error {
	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(ref.Pod).
		Namespace(ref.Namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: ref.Container,
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
	if err != nil {
		return err
	}

	return nil
}

// WaitForTerminal is called from apihandler.handleAttach as a goroutine
// Waits for the SockJS connection to be opened by the client the session to be bound in handleTerminalSession
func WaitForTerminal(k8sClient kubernetes.Interface, cfg *rest.Config, ref ContainerRef) {
	<-terminalSessions.Get(ref.String()).bound
	close(terminalSessions.Get(ref.String()).bound)

	var err error
	validShells := []string{"bash", "sh"}

	for _, testShell := range validShells {
		cmd := []string{testShell}
		if err = startProcess(k8sClient, cfg, ref, cmd, terminalSessions.Get(ref.String())); err == nil {
			break
		}
	}

	if err != nil {
		terminalSessions.Close(ref.String(), 2, err.Error())
		return
	}

	terminalSessions.Close(ref.String(), 1, "Process exited")
}

type ContainerRef struct {
	Environment string
	Cluster     string
	ClusterID   uint
	Namespace   string
	Pod         string
	Container   string
	RandomID    string
}

func (ref ContainerRef) String() string {
	return genSessionID(ref.ClusterID, ref.Pod, ref.Container, ref.RandomID)
}

func ParseContainerRef(str string) (ContainerRef, error) {
	strs := strings.Split(str, ":")
	if len(strs) != 6 {
		return ContainerRef{}, errors.New("invalid container ref")
	}

	return ContainerRef{
		Environment: strs[0],
		Cluster:     strs[1],
		Namespace:   strs[2],
		Pod:         strs[3],
		Container:   strs[4],
		RandomID:    strs[5],
	}, nil
}
