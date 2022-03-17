# Error Practice in Horizon

## Do Not Throw Third Party Error Directly

Before throwing error to its caller, it should be wrapped with more information,
such as stack info, error msg and operation. That will help developers localize bugs.

```go
package "g.hz.netease.com/horizon/pkg/gitrepo" 
import (
    "g.hz.netease.com/horizon/pkg/errors"
)
var (
	GitNotFoundErr = errors.New("Not Found")
)
type HorizonGitRepoInterface interface {
}

type HorizonGitRepoInterfaceImp struct {
	GitLabClient *gitlabExternalCall
}


func (h *HorizonGitInterfaceImp) funca(ctx context.Context, file string)(interface{}, error) {
	_, resp, err:= h.GitLabClient.Call(file)
	if resp.StatusCode == http.StatusNotFound {
	  return nil, errors.Wrap(GitNotFoundErr, err.Error())	
    }
	
	return nil, nil
}
```

## Use Errors in Horizon

Horizon defines all errors in a file which you could find it in [horizonerrors.go](../../core/errors/horizonerrors.go).
There are two ways to define a kind of error. First, define an error type, such as 

```go
type HorizonErrNotFound struct {
	Source sourceType
}

func NewErrNotFound(source sourceType, msg string) error {
    return errors.Wrap(&HorizonErrNotFound{
        Source: source,
    }, msg)
}

func (e *HorizonErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", e.Source.name)
}
```

`HorizonErrNotFound` has a field called `Source`, it shows which resource leads the error. 
Horizon decouples error type and resource, for there's many errors in project are associated with resource,
if defining an error for every error type and every source, there'll be too much redundancy.
Second, for errors not related to resource, Horizon defines errors directly, for example

```go
ErrParamInvalid = errors.New("parameter is invalid")
```

This is quite simple. Note that, it should be wrapped manually, likes `HorizonErrNotFound` does in `NewErrNotFound`

```go
return perror.Wrap(herrors.ErrParamInvalid, "application config for template cannot be empty")
```

On the above, Horizon uses `perror.Wrap(err)` getting the underlying error.
Correspondingly, there's two ways to handle errors.
For error types like `HorizonErrNotFound`, handling it with

```go
if _, ok := perror.Cause(err).(*herrors.NewErrNotFound); ok {
	...
}
```

For errors defined directly, handling it like this

```go
if perror.Wrap(err) == herrors.ErrParamInvalid {
	...
}
```

### 3. Print Errors

For Horizon wrapped error with stack info and message, in same cases, you don't want to see stack info,
 and print it like this

```go
func foo() error { 
	return errors.Wrap(sql.ErrNoRows, "foo failed")
}

func bar() error { 
	return errors.WithMessage(foo(), "bar failed")
}

func main() {
	err := bar()
    fmt.Printf("data not found, %v\n", err)
    // Output: 
    // bar failed: foo failed: sql: no rows in result set 
}
```

Print with stack info

```go
func main() {
    err := bar()
    fmt.Printf("%+v\n", err)
	// Output:
	// sql: no rows in result set
    // foo failed
    // main.foo
    // /usr/three/main.go:11
    // main.bar
    // /usr/three/main.go:15
    // main.main
    // /usr/three/main.go:19
    // runtime.main
    // ... 
}
```
