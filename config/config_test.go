package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kudrykv/go-vkpm/config"
	. "github.com/smartystreets/goconvey/convey"
)

const testConfig = `
domain: domain
default_project: defproj
cookies:
  csrftoken: csrf
  sessionid: sessid
`

func TestNew(t *testing.T) {
	Convey("New", t, func() {
		Convey("brand new", func() {
			_ = os.Remove("./test_config.yml")

			cfg, err := config.New(".", "test_config.yml")
			So(err, ShouldBeNil)

			cfg.Domain = "domain"
			cfg.DefaultProject = "defproj"
			cfg.Cookies = config.Cookies{CSRFToken: "csrf", SessionID: "sessid"}
			So(cfg.Write(), ShouldBeNil)

			read, err := cfg.Read()
			So(err, ShouldBeNil)
			So(read, ShouldResemble, cfg)
		})

		Convey("existing", func() {
			err := ioutil.WriteFile("./test_existing_config.yml", []byte(testConfig), 0600)
			So(err, ShouldBeNil)

			cfg, err := config.New(".", "test_existing_config.yml")
			So(err, ShouldBeNil)

			expected := cfg // need to copy, as has internal fields. it is easier this way
			expected.Domain = "domain"
			expected.DefaultProject = "defproj"
			expected.Cookies = config.Cookies{CSRFToken: "csrf", SessionID: "sessid"}
			So(cfg, ShouldResemble, expected)
		})
	})
}
