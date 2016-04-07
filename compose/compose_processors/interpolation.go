package compose_processors

import (
	"github.com/vmware/harbor/compose/compose"
	"log"
	"regexp"
	"strings"
)

var (
	INTERPOLATION_PATTERN_0 = regexp.MustCompile(`\$\{[\w|_|-]+\}`)
	INTERPOLATION_PATTERN_1 = regexp.MustCompile(`\$[\w|_|-]+`)
)

func init() {
	Processors = append(Processors, Interpolation)
}

func Interpolation(compose *compose.SryCompose) *compose.SryCompose {
	for _, app := range compose.Applications {
		_interpolation(&app.Name, compose.Answers)
		_interpolation(&app.Image, compose.Answers)
		_interpolation(&app.Net, compose.Answers)
		_interpolation(&app.EntryPoint, compose.Answers)

		for _, volume := range app.Volumes {
			_interpolation(&volume.Host, compose.Answers)
			_interpolation(&volume.Container, compose.Answers)
		}

		for _, env := range app.Environment {
			_interpolation(&env.Value, compose.Answers)
		}

		for _, label := range app.Labels {
			_interpolation(&label.Value, compose.Answers)
		}
	}

	for _, app := range compose.Applications {
		log.Println(app.Image)
	}

	return compose
}

func _interpolation(str *string, answers map[string]string) {
	replaceFunc := func(matched []byte) []byte {
		env := strings.ToLower(string(matched[2 : len(matched)-1]))
		if v, ok := answers[env]; ok {
			return []byte(v)
		}
		return matched
	}
	*str = string(INTERPOLATION_PATTERN_0.ReplaceAllFunc([]byte(*str), replaceFunc))

	replaceFunc1 := func(matched []byte) []byte {
		env := strings.ToLower(string(matched[1:]))
		if v, ok := answers[env]; ok {
			return []byte(v)
		}
		return matched
	}
	*str = string(INTERPOLATION_PATTERN_1.ReplaceAllFunc([]byte(*str), replaceFunc1))
}
