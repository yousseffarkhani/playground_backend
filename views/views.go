package views

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/yousseffarkhani/playground/backend2/server"
)

// TODO : Put back when main is in /cmd file
// const layoutDir = "../views/layouts"
// const templateDir = "../views/templates"
const layoutDir = "views/layouts"
const templateDir = "views/templates"

type View struct {
	template *template.Template
	Layout   string
}

func Initialize() map[string]server.View {
	var views = make(map[string]server.View)
	views["home"] = newView("main", templateDir+"/home.html")
	views["playgrounds"] = newView("main", templateDir+"/playgrounds.html")
	views["playground"] = newView("main", templateDir+"/playground.html")
	views["login"] = newView("main", templateDir+"/login.html")
	views["404"] = newView("main", templateDir+"/404.html")
	views["internal error"] = newView("main", templateDir+"/internalError.html")
	return views
}

func newView(layout string, files ...string) *View {
	files = append(files, layoutFiles()...)
	tmpl := template.Must(template.ParseFiles(files...))
	return &View{
		template: tmpl,
		Layout:   layout,
	}
}

func layoutFiles() []string {
	files, err := filepath.Glob(layoutDir + "/*.html")
	if err != nil {
		fmt.Println(err)
	}
	return files
}

func (v *View) Render(w io.Writer, r *http.Request, data server.RenderingData) error {
	err := v.template.ExecuteTemplate(w, v.Layout, data)
	return err
}
