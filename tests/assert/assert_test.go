package assert_test

import (
	"errors"
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

// recordingTB stands in for a *testing.T so the failure reporters can be
// called without failing the test that calls them. Only Fail and FailNow are
// reached; the embedded nil TB is never used.
type recordingTB struct {
	testing.TB

	failed bool
}

func (r *recordingTB) Fail() {
	r.failed = true
}

func (r *recordingTB) FailNow() {
	r.failed = true
}

// suiteFixture exercises TestSuite and Run.
type suiteFixture struct {
	assert.TestSuite

	setup    int
	tearDown int
	ran      int
}

func (s *suiteFixture) SetupTest() {
	s.setup++
}

func (s *suiteFixture) TearDownTest() {
	s.tearDown++
}

func (s *suiteFixture) TestPasses() {
	s.ran++
	assert.True(s.T(), true, "the suite hands its T to the test")
}

// TestAssert invokes every assertion on the path where it holds.
func TestAssert(t *testing.T) {
	t.Run("core", assertCore)
	t.Run("reporters", assertReporters)
	t.Run("suite", assertSuite)
	t.Run("extra", assertExtra)
}

func assertCore(t *testing.T) {
	assert.True(t, assert.ObjectsAreEqualValues(1, 1))
	assert.False(t, assert.ObjectsAreEqualValues(1, 2))

	assert.Equal(t, "a", "a", "strings match")
	assert.EqualValues(t, []int{1, 2}, []int{1, 2})
	assert.NotEqual(t, 1, 2)

	assert.NoError(t, nil)
	assert.Error(t, errors.New("boom"))

	assert.True(t, true)
	assert.False(t, false)

	assert.Contains(t, "haystack", "stack")
	assert.Contains(t, []string{"a", "b"}, "b")
	assert.Contains(t, map[string]int{"a": 1}, "a")

	assert.CheckEquals(t, 42, 42, "integers match")
	assert.Assert(t, true, "condition holds for %d", 42)
}

// assertReporters drives Errorf and Fail against a stand-in TB, so the failure
// they raise is recorded instead of failing this test. Both print the message
// to stdout, which is what the two lines in a passing run are.
func assertReporters(t *testing.T) {
	errorf := &recordingTB{}
	assert.Errorf(errorf, "reporter check: %d", 42)
	assert.True(t, errorf.failed, "Errorf fails the TB it is given")

	fail := &recordingTB{}
	assert.Fail(fail, "reporter check", "fail")
	assert.True(t, fail.failed, "Fail fails the TB it is given")
}

func assertSuite(t *testing.T) {
	fixture := &suiteFixture{}
	fixture.SetT(t)
	assert.True(t, fixture.T() == t, "SetT and T agree")

	assert.Run(t, fixture)
	assert.Equal(t, 1, fixture.setup)
	assert.Equal(t, 1, fixture.ran)
	assert.Equal(t, 1, fixture.tearDown)
}
