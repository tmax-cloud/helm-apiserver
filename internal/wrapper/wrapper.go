package wrapper

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

// RouterWrapper is an interface for wrapper
type RouterWrapper interface {
	Add(child RouterWrapper) error
	FullPath() string

	Router() *mux.Router
	SetRouter(*mux.Router)

	Children() []RouterWrapper

	Parent() RouterWrapper
	SetParent(RouterWrapper)

	Handler() http.HandlerFunc
	SubPath() string
	Methods() []string
}

// wrapper wraps router with tree structure
type wrapper struct {
	router *mux.Router

	subPath string
	methods []string
	handler http.HandlerFunc

	children []RouterWrapper
	parent   RouterWrapper
}

// New is a constructor for the wrapper
func New(path string, methods []string, handler http.HandlerFunc) *wrapper {
	return &wrapper{
		subPath: path,
		methods: methods,
		handler: handler,
	}
}

// Router returns its router
func (w *wrapper) Router() *mux.Router {
	return w.router
}

// SetRouter sets its router
func (w *wrapper) SetRouter(r *mux.Router) {
	w.router = r
}

// Children returns its children
func (w *wrapper) Children() []RouterWrapper {
	return w.children
}

// Parent returns its parent
func (w *wrapper) Parent() RouterWrapper {
	return w.parent
}

// SetParent sets parent
func (w *wrapper) SetParent(parent RouterWrapper) {
	w.parent = parent
}

// Handler returns its handler
func (w *wrapper) Handler() http.HandlerFunc {
	return w.handler
}

// SubPath returns its subPath
func (w *wrapper) SubPath() string {
	return w.subPath
}

// Methods returns its methods
func (w *wrapper) Methods() []string {
	return w.methods
}

// Add adds child as a child (child node of a tree) of w
func (w *wrapper) Add(child RouterWrapper) error {
	if child == nil || child.(*wrapper) == nil {
		return fmt.Errorf("child is nil")
	}

	if child.Parent() != nil {
		return fmt.Errorf("child already has parent")
	}

	if child.SubPath() == "" || child.SubPath() == "/" || child.SubPath()[0] != '/' {
		return fmt.Errorf("child subpath is not valid")
	}

	if w.router == nil {
		return fmt.Errorf("parent does not have a router")
	}

	child.SetParent(w)
	w.children = append(w.children, child)

	child.SetRouter(w.router.PathPrefix(child.SubPath()).Subrouter())

	if child.Handler() != nil {
		if len(child.Methods()) > 0 {
			child.Router().Methods(child.Methods()...).Subrouter().HandleFunc("/", child.Handler())
			w.router.Methods(child.Methods()...).Subrouter().HandleFunc(child.SubPath(), child.Handler())
		} else {
			child.Router().HandleFunc("/", child.Handler())
			w.router.HandleFunc(child.SubPath(), child.Handler())
		}
	}

	return nil
}

// FullPath builds full path string of the api
func (w *wrapper) FullPath() string {
	if w.parent == nil {
		return w.subPath
	}
	re := regexp.MustCompile(`/{2,}`)
	return re.ReplaceAllString(w.parent.FullPath()+w.subPath, "/")
}
