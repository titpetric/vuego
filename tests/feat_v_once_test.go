package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/helpers"
)

func TestVOnce_SkipsSubsequentRenders(t *testing.T) {
	const template = "v-once-test.vuego"

	// Template with v-once inside v-for loop
	templateFS := &fstest.MapFS{
		template: {Data: []byte(`<div v-for="item in items">
  <script v-once>
    (function () {
      let appearance = localStorage.getItem("appearance");
      appearance && document.documentElement.setAttribute("data-appearance", appearance);
    })();
  </script>
  <p>{{ item }}</p>
</div>`)},
	}

	data := map[string]any{
		"items": []string{"apple", "banana", "cherry"},
	}

	// Expected output: script tag only appears once, even though v-for renders 3 items
	want := []byte(`<div>
  <script>
    (function () {
      let appearance = localStorage.getItem("appearance");
      appearance && document.documentElement.setAttribute("data-appearance", appearance);
    })();
  </script>
  <p>apple</p>
</div>
<div>
  <p>banana</p>
</div>
<div>
  <p>cherry</p>
</div>
`)

	vue := vuego.NewVue(templateFS)
	var got bytes.Buffer
	require.NoError(t, vue.RenderFragment(&got, template, data))
	require.True(t, helpers.EqualHTML(t, want, got.Bytes(), nil, nil))

	t.Logf("-- v-once result: %s", got.String())
}

func TestVOnce_MultipleVOnceElements(t *testing.T) {
	const template = "v-once-multi.vuego"

	// Template with multiple v-once elements inside v-for
	templateFS := &fstest.MapFS{
		template: {Data: []byte(`<section v-for="item in items">
  <script v-once src="/assets/js/component.js"></script>
  <style v-once>
    .component { color: blue; }
  </style>
  <span>{{ item }}</span>
</section>`)},
	}

	data := map[string]any{
		"items": []string{"first", "second", "third"},
	}

	// Expected: each v-once element appears only once
	want := []byte(`<section>
  <script src="/assets/js/component.js"></script>
  <style>
    .component { color: blue; }
  </style>
  <span>first</span>
</section>
<section>
  <span>second</span>
</section>
<section>
  <span>third</span>
</section>
`)

	vue := vuego.NewVue(templateFS)
	var got bytes.Buffer
	require.NoError(t, vue.RenderFragment(&got, template, data))
	require.True(t, helpers.EqualHTML(t, want, got.Bytes(), nil, nil))

	t.Logf("-- v-once multiple result: %s", got.String())
}
