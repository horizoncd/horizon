## About Error Best Practice



### 1. do not throw thirdparty  error direct（Wrap with horizon error）
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


func (h *HorizonGitInterfaceImp) funca(ctx context.Context, file string)( interface{},error) {
	_, resp, err:= h.GitLabClient.Call(file)
	if resp.StatusCode == http.StatusNotFound {
	  return ,nil errors.Wrap(GitNotFoundErr, err.Error())	
    }
}
```


### 2. use horizon error in horizon project
```go
import (
	"g.hz.netease.com/horizon/pkg/gitrepo"
    "g.hz.netease.com/horizon/pkg/errors"
)

type HorizonModuleA interface {
}

type HorizonGitInterfaceImp struct {
    gitRepo *HorizonGitInterface
}

func (h *HorizonGitInterfaceImp) funcb(ctx context.Context, file string) ( interface{}, error) {
	file, err := gitRepo.funca()
	
	// case 1 (you care about the sepcial error),
	// you can break the error stack(as case 1)
	// or you can attach some message and return
	if errors.Case(err) == gitrepo.GitNotFoundErr {
	    // do some thing
    }  else {
    	return "", errors.WithMessagef(err, "gitrepo funca return error, err = %s",err.Error())
    }   
    
    
    // case 2 (all error process logical is same, just return error)
    if err != nil {
    	return "", err
    }
}
```
