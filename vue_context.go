package vuego

import (
	"context"
	"path"
	"reflect"
	"strings"

	"github.com/titpetric/vuego/internal/ulid"
)

// VueContext carries template inclusion context and request-scoped state used during evaluation.
// Each render operation gets its own VueContext, making concurrent rendering safe.
type VueContext struct {
	// Variable scope and data resolution
	ctx   context.Context
	stack *Stack

	// BaseDir is the root directory for template inclusion chains.
	BaseDir string
	// CurrentDir is the current working directory during template processing.
	CurrentDir string
	// FromFilename is the name of the file currently being processed.
	FromFilename string
	// TemplateStack is the stack of included template files.
	TemplateStack []string

	// TagStack is the stack of HTML tags being rendered.
	TagStack []string

	// Processors are the registered template processors.
	Processors []NodeProcessor

	// v-once element tracking for deep clones
	seen map[string]bool

	// SlotScope contains slot content for the current component.
	SlotScope *SlotScope
}

// VueContextOptions holds configurable options for a new VueContext.
type VueContextOptions struct {
	// Stack is the resolver stack for variable lookups.
	Stack *Stack
	// Processors are the registered template processors.
	Processors []NodeProcessor
}

// NewVueContext returns a VueContext initialized for the given template filename with initial data.
func NewVueContext(ctx context.Context, fromFilename string, options *VueContextOptions) VueContext {
	result := VueContext{
		ctx:           ctx,
		stack:         options.Stack,
		CurrentDir:    path.Dir(fromFilename),
		FromFilename:  fromFilename,
		TemplateStack: []string{fromFilename},
		TagStack:      []string{},
		seen:          make(map[string]bool),
	}
	for _, v := range options.Processors {
		result.Processors = append(result.Processors, v.New())
	}
	return result
}

// WithTemplate returns a copy of the context extended with filename in the inclusion chain.
func (ctx VueContext) WithTemplate(filename string) VueContext {
	newStack := make([]string, len(ctx.TemplateStack)+1)
	copy(newStack, ctx.TemplateStack)
	newStack[len(ctx.TemplateStack)] = filename
	return VueContext{
		ctx:           ctx.ctx,
		stack:         ctx.stack, // Share the same stack across template chain
		BaseDir:       ctx.BaseDir,
		CurrentDir:    path.Dir(filename),
		FromFilename:  filename,
		TemplateStack: newStack,
		TagStack:      ctx.TagStack, // Share the same tag stack
		seen:          ctx.seen,     // Share the v-once tracking map
		Processors:    ctx.Processors,
		SlotScope:     ctx.SlotScope, // Share the slot scope
	}
}

// FormatTemplateChain returns the template inclusion chain formatted for error messages.
func (ctx VueContext) FormatTemplateChain() string {
	if len(ctx.TemplateStack) <= 1 {
		return ctx.FromFilename
	}
	return strings.Join(ctx.TemplateStack, " -> ")
}

// PushTag adds a tag to the tag stack.
func (ctx *VueContext) PushTag(tag string) {
	ctx.TagStack = append(ctx.TagStack, tag)
}

// PopTag removes the current tag from the tag stack.
func (ctx *VueContext) PopTag() {
	if len(ctx.TagStack) > 0 {
		ctx.TagStack = ctx.TagStack[:len(ctx.TagStack)-1]
	}
}

// CurrentTag returns the current parent tag, or empty string if no tag is on the stack.
func (ctx VueContext) CurrentTag() string {
	if len(ctx.TagStack) == 0 {
		return ""
	}
	return ctx.TagStack[len(ctx.TagStack)-1]
}

// Context returns the context.Context associated with this VueContext.
func (ctx VueContext) Context() context.Context {
	return ctx.ctx
}

// ExprEnv returns the env map for expr evaluation, wrapping any functions whose
// first parameter is context.Context so the context is injected automatically.
func (ctx VueContext) ExprEnv() map[string]any {
	env := ctx.stack.EnvMap()
	ctx.bindContextFuncs(env)
	return env
}

// bindContextFuncs wraps function values in env whose first parameter is
// context.Context, binding ctx.ctx so expr can call them without arguments.
func (ctx VueContext) bindContextFuncs(env map[string]any) {
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	for k, v := range env {
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Func {
			continue
		}
		ft := rv.Type()
		if ft.NumIn() == 0 {
			continue
		}
		if ft.In(0) != contextType {
			continue
		}
		env[k] = ctx.wrapContextFunc(rv, ft)
	}
}

// wrapContextFunc returns a new function that prepends ctx.ctx to the call.
func (ctx VueContext) wrapContextFunc(fn reflect.Value, ft reflect.Type) any {
	// Build the wrapped function type: same signature minus the first context.Context param.
	numIn := ft.NumIn()
	inTypes := make([]reflect.Type, numIn-1)
	for i := 1; i < numIn; i++ {
		inTypes[i-1] = ft.In(i)
	}
	outTypes := make([]reflect.Type, ft.NumOut())
	for i := 0; i < ft.NumOut(); i++ {
		outTypes[i] = ft.Out(i)
	}
	wrappedType := reflect.FuncOf(inTypes, outTypes, ft.IsVariadic())
	wrapped := reflect.MakeFunc(wrappedType, func(args []reflect.Value) []reflect.Value {
		fullArgs := make([]reflect.Value, 0, len(args)+1)
		fullArgs = append(fullArgs, reflect.ValueOf(ctx.ctx))
		fullArgs = append(fullArgs, args...)
		return fn.Call(fullArgs)
	})
	return wrapped.Interface()
}

// Stack returns the variable resolution stack for this context.
// This allows functions with *VueContext parameters to resolve variables from the execution scope.
func (ctx VueContext) Stack() *Stack {
	return ctx.stack
}

// nextSeenID returns a unique ID for tracking v-once elements across deep clones.
func (ctx *VueContext) nextSeenID() string {
	return ulid.String()
}
