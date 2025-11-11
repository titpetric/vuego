package vuego_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/titpetric/vuego"
)

var engines = []vuego.Engine{
	vuego.New(),
}

func TestRenderEngine(t *testing.T) {
	const template = `<app><h1 v-for="section in sections" v-html="section.title"></h1></app>`

	for _, engine := range engines {
		vuego.RenderEngine(os.Stdout, engine, template, map[string]any{
			"sections": []map[string]any{
				{
					"title": "Section 1",
				},
				{
					"title": "Section 2",
				},
			},
		})
	}
}

func BenchmarkRenderEngine(b *testing.B) {
	const template = `<app><h1 v-for="section in sections" v-html="section.title"></h1></app>`

	for _, engine := range engines {

		b.Run(fmt.Sprintf("%T", engine), func(b *testing.B) {
			for _ = range b.N {
				var w bytes.Buffer
				vuego.RenderEngine(&w, engine, template, map[string]any{
					"sections": []map[string]any{
						{
							"title": "Section 1",
						},
						{
							"title": "Section 2",
						},
					},
				})
			}
		})
	}
}

func TestRenderEngine_Complex(t *testing.T) {
	const template = `
<div>
  <ul>
    <li v-for="(i, item) in items" :title="item.title">
      <h1 v-html="item.title"></h1>
      <span>index: {i}</span>
    </li>
  </ul>
</div>
`

	for _, engine := range engines {
		data := map[string]any{
			"items": []map[string]any{
				{"title": "One"},
				{"title": "Two"},
				{"title": "Three"},
			},
		}
		vuego.RenderEngine(os.Stdout, engine, template, data)
	}
}
