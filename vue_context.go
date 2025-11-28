package vuego

import (
	"fmt"
	"path"
	"strings"
)

// VueContext carries template inclusion context and request-scoped state used during evaluation.
// Each render operation gets its own VueContext, making concurrent rendering safe.
type VueContext struct {
	// Variable scope and data resolution
	stack *Stack

	// Template inclusion chain context
	BaseDir       string
	CurrentDir    string
	FromFilename  string
	TemplateStack []string

	// HTML rendering state
	TagStack []string

	Processors []NodeProcessor

	// v-once element tracking for deep clones
	seen        map[string]bool
	seenCounter int
}

// VueContextOptions holds configurable options for a new VueContext.
type VueContextOptions struct {
	Data         map[string]any
	OriginalData any
	Processors   []NodeProcessor
}

// NewVueContext returns a VueContext initialized for the given template filename with initial data.
func NewVueContext(fromFilename string, options *VueContextOptions) VueContext {
	result := VueContext{
		stack:         NewStackWithData(options.Data, options.OriginalData),
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
		stack:         ctx.stack, // Share the same stack across template chain
		BaseDir:       ctx.BaseDir,
		CurrentDir:    path.Dir(filename),
		FromFilename:  filename,
		TemplateStack: newStack,
		TagStack:      ctx.TagStack, // Share the same tag stack
		seen:          ctx.seen,     // Share the v-once tracking map
		Processors:    ctx.Processors,
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

// nextSeenID returns a unique ID for tracking v-once elements across deep clones.
func (ctx *VueContext) nextSeenID() string {
	ctx.seenCounter++
	return fmt.Sprint(ctx.seenCounter)
}
