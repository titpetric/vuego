package vuego_test

import (
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

// Sample model types the typed-passthrough table binds.
type (
	Article struct {
		Title string
		Body  string
	}

	BlogPost struct {
		ID        int
		Title     string
		Content   string
		Tags      []string
		Published bool
	}
)

func TestVue_EvalAttributes(t *testing.T) {
	tests := []struct {
		name   string
		source string
		data   map[string]any
		want   string
	}{
		// Binding syntax.
		{
			name:   "colon prefix",
			source: `<div :id="id"></div>`,
			data:   map[string]any{"id": "test-div"},
			want:   `<div id="test-div"></div>`,
		},
		{
			name:   "v-bind prefix",
			source: `<div v-bind:class="class"></div>`,
			data:   map[string]any{"class": "active"},
			want:   `<div class="active"></div>`,
		},
		{
			name:   "brackets escape the binding syntax",
			source: `<div [:id]="id"></div>`,
			data:   map[string]any{"id": "test-div"},
			want:   `<div :id="id"></div>`,
		},
		{
			name:   "brackets keep the attribute but still interpolate",
			source: `<div [:field]="{{ field | json }}"></div>`,
			data:   map[string]any{"field": map[string]any{"id": "test-div"}},
			want:   `<div :field="{&#34;id&#34;:&#34;test-div&#34;}"></div>`,
		},

		// Truthiness decides whether a bound attribute is written at all.
		{
			name:   "false drops the attribute",
			source: `<button :disabled="disabled"></button>`,
			data:   map[string]any{"disabled": false},
			want:   `<button></button>`,
		},
		{
			name:   "true writes the attribute",
			source: `<button :disabled="disabled"></button>`,
			data:   map[string]any{"disabled": true},
			want:   `<button disabled="true"></button>`,
		},
		{
			name:   "an empty string drops the attribute",
			source: `<input :required="value" />`,
			data:   map[string]any{"value": ""},
			want:   `<input />`,
		},
		{
			name:   "zero drops the attribute",
			source: `<span :data-count="counter"></span>`,
			data:   map[string]any{"counter": 0},
			want:   `<span></span>`,
		},
		{
			name:   "a non-zero number writes the attribute",
			source: `<span :data-count="counter"></span>`,
			data:   map[string]any{"counter": 5},
			want:   `<span data-count="5"></span>`,
		},
		{
			name:   "a missing variable drops the attribute",
			source: `<div :title="missing"></div>`,
			data:   map[string]any{},
			want:   `<div></div>`,
		},

		// Static attributes and interpolation.
		{
			name:   "a static attribute is left alone",
			source: `<div class="static"></div>`,
			data:   map[string]any{},
			want:   `<div class="static"></div>`,
		},
		{
			name:   "a static attribute is interpolated",
			source: `<div class="item-{{ id }}"></div>`,
			data:   map[string]any{"id": "123"},
			want:   `<div class="item-123"></div>`,
		},
		{
			name:   "interpolation can sit between literals",
			source: `<div data-info="prefix-{{ value }}-suffix"></div>`,
			data:   map[string]any{"value": "middle"},
			want:   `<div data-info="prefix-middle-suffix"></div>`,
		},

		// A bound value is used as-is: its contents are not re-evaluated.
		{
			name:   "a bound value holding mustaches is not interpolated",
			source: `<div :title="content"></div>`,
			data:   map[string]any{"content": "{{ name }}", "name": "Alice"},
			want:   `<div title="{{ name }}"></div>`,
		},
		{
			name:   "a bound value holding a pipe is not filtered",
			source: `<div :title="text"></div>`,
			data:   map[string]any{"text": "a | b"},
			want:   `<div title="a | b"></div>`,
		},
		{
			name:   "a bound value holding shell syntax is escaped, not evaluated",
			source: `<span :data-code="code"></span>`,
			data:   map[string]any{"code": `echo "${{ version }}"`},
			want:   `<span data-code="echo &#34;${{ version }}&#34;"></span>`,
		},
		{
			name:   "bound and interpolated agree for a plain string",
			source: `<div :title="msg" data-alt="{{ msg }}"></div>`,
			data:   map[string]any{"msg": "hello"},
			want:   `<div title="hello" data-alt="hello"></div>`,
		},

		// Several bindings on one element.
		{
			name:   "several bound attributes",
			source: `<button :id="id" :class="btnClass" :disabled="disabled"></button>`,
			data:   map[string]any{"id": "submit-btn", "btnClass": "primary", "disabled": false},
			want:   `<button id="submit-btn" class="primary"></button>`,
		},
		{
			name:   "bound attributes beside static ones",
			source: `<input type="text" :value="name" :placeholder="hint" class="form-input" />`,
			data:   map[string]any{"name": "John", "hint": "Enter your name"},
			want:   `<input type="text" value="John" placeholder="Enter your name" class="form-input" />`,
		},
		{
			name:   "several v-bind attributes",
			source: `<a v-bind:href="url" v-bind:target="target" v-bind:title="title">Link</a>`,
			data:   map[string]any{"url": "/page", "target": "_blank", "title": "External"},
			want:   `<a href="/page" target="_blank" title="External">Link</a>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertRenderHTML(t, test.source, test.data, test.want)
		})
	}
}

func TestVue_EvalBoundAttribute(t *testing.T) {
	tests := []struct {
		name   string
		source string
		data   map[string]any
		want   string
	}{
		{
			name:   "a variable reference",
			source: `<div :data-value="id"></div>`,
			data:   map[string]any{"id": "123"},
			want:   `<div data-value="123"></div>`,
		},
		{
			name:   "a path reference",
			source: `<a :href="post.url"></a>`,
			data:   map[string]any{"post": map[string]any{"url": "/posts/42"}},
			want:   `<a href="/posts/42"></a>`,
		},
		{
			name:   "a function call",
			source: `<div :title="len(items)"></div>`,
			data:   map[string]any{"items": []int{1, 2, 3}},
			want:   `<div title="3"></div>`,
		},
		{
			name:   "a nested function call",
			source: `<p :data-text="trim(text)"></p>`,
			data:   map[string]any{"text": "  spaces  "},
			want:   `<p data-text="spaces"></p>`,
		},
		{
			name:   "a filter with an argument",
			source: `<div :title="message | default('no title')"></div>`,
			data:   map[string]any{},
			want:   `<div title="no title"></div>`,
		},
		{
			name:   "a filter on a missing variable",
			source: `<span :data-value="missing | default('fallback')"></span>`,
			data:   map[string]any{},
			want:   `<span data-value="fallback"></span>`,
		},
		{
			name:   "a comparison expression",
			source: `<div :data-big="count > 2"></div>`,
			data:   map[string]any{"count": 5},
			want:   `<div data-big="true"></div>`,
		},
		{
			name:   "an interpolated bound value",
			source: `<div :title="Hi {{ name }}"></div>`,
			data:   map[string]any{"name": "Alice"},
			want:   `<div title="Hi Alice"></div>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertRenderHTML(t, test.source, test.data, test.want)
		})
	}
}

func TestVue_TypedAttributePassthrough(t *testing.T) {
	// A bound attribute keeps the Go type of its value across a template
	// include, so filters like type and len see the original value.
	tests := []struct {
		name     string
		parent   string
		child    string
		data     map[string]any
		contains []string
	}{
		{
			name:     "slice of structs",
			parent:   `<template include="child.vuego" :articles="articles"></template>`,
			child:    `<template>{{ articles | type }}</template>`,
			data:     map[string]any{"articles": []Article{{Title: "Article 1"}, {Title: "Article 2"}}},
			contains: []string{"[]vuego_test.Article"},
		},
		{
			name:     "struct value",
			parent:   `<template include="child.vuego" :article="article"></template>`,
			child:    `<template>{{ article | type }}</template>`,
			data:     map[string]any{"article": Article{Title: "Test", Body: "Test body"}},
			contains: []string{"vuego_test.Article"},
		},
		{
			name:     "map",
			parent:   `<template include="child.vuego" :data="myMap"></template>`,
			child:    `<template>{{ data | type }}</template>`,
			data:     map[string]any{"myMap": map[string]string{"key": "value"}},
			contains: []string{"map[string]string"},
		},
		{
			name:     "int",
			parent:   `<template include="child.vuego" :count="count"></template>`,
			child:    `<template>{{ count | type }}</template>`,
			data:     map[string]any{"count": 42},
			contains: []string{"int"},
		},
		{
			name:   "slice of structs stays countable",
			parent: `<template include="child.vuego" :posts="posts"></template>`,
			child:  "<template>\nType: {{ posts | type }}\nCount: {{ posts | len }}\n</template>",
			data: map[string]any{"posts": []BlogPost{
				{ID: 1, Title: "Post 1", Tags: []string{"go", "testing"}},
				{ID: 2, Title: "Post 2", Tags: []string{"go", "types"}},
			}},
			contains: []string{"Type: []vuego_test.BlogPost", "Count: 2"},
		},
		{
			name:     "string slice stays countable",
			parent:   `<template include="child.vuego" :tags="tags"></template>`,
			child:    "<template>\nType: {{ tags | type }}\nCount: {{ tags | len }}\n</template>",
			data:     map[string]any{"tags": []string{"go", "template", "testing"}},
			contains: []string{"Type: []string", "Count: 3"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := renderInclude(t, test.parent, map[string]string{"child.vuego": test.child}, test.data)
			for _, want := range test.contains {
				assert.Contains(t, got, want, "output: %s", got)
			}
		})
	}
}
