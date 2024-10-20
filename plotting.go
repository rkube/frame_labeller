package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Renders an echart using a html template, see
// https://blog.cubieserver.de/2020/how-to-render-standalone-html-snippets-with-go-echarts/

// type Renderer interface {
// 	Render(w io.Writer) error
// }

// type snippetRenderer struct {
// 	c      interface{}
// 	before []func()
// }

// func newSnippetRenderer(c interface{}, before ...func()) chartrender.Renderer {
// 	return &snippetRenderer{c: c, before: before}
// }

// func (r *snippetRenderer) Render(w io.Writer) error {
// 	const tplName = "chart"
// 	for _, fn := range r.before {
// 		fn()
// 	}

// 	tpl := template.
// 		Must(template.New(tplName).
// 			Funcs(template.FuncMap{
// 				"safeJS": func(s interface{}) template.JS {
// 					return template.JS(fmt.Sprint(s))
// 				},
// 			}).
// 			Parse(baseTpl),
// 		)

// 	err := tpl.ExecuteTemplate(w, tplName, r.c)
// 	return err
// }

// func renderToHtml(c interface{}) template.HTML {
// 	var buf bytes.Buffer
// 	// r := c.(chartrender.Renderer)
// 	r := c.(chartrender.RenderSnippets)
// 	err := r.Render(&buf)
// 	if err != nil {
// 		log.Printf("Failed to render chart: %s", err)
// 		return ""
// 	}

// 	return template.HTML(buf.String())
// }

// func render_html_snippet

func fetch_sparta_plot(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	// Parse the form and find the requested frame
	err := r.ParseForm()
	if err != nil { // Return a Bad Request if we can't parse the form
		fmt.Fprintf(os.Stdout, "signin_handler: Unable to parse %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return 400, nil
	}
	fmt.Println("fetch_sparta_plot: range= ", r.Form["range"][0])

	// Get the current frame from the request
	current_frame := -1
	if key, ok := r.Form["range"]; ok {
		val, err := strconv.Atoi(key[0])
		if err != nil {
			return 0, err
		}
		current_frame = val
	}
	fmt.Println("Fetching data for frame ", current_frame)

	// If the user is logged in, update the shot the user is on
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("fetch_data_array: Session token not set")
	} else {
		// session_id = c.Value
		username := a.session_to_user[c.Value]
		fmt.Println("fetch_sparta_plot: username = ", username)
	}

	// Create a new chart and put it in an html template
	// pie := charts.NewPie()

	// // preformat data
	// pieData := []opts.PieData{
	// 	{Name: "Dead Cases", Value: 123},
	// 	{Name: "Recovered Cases", Value: 456},
	// 	{Name: "Active Cases", Value: 789},
	// }

	// // put data into chart
	// pie.AddSeries("Case Distribution", pieData).SetSeriesOptions(
	// 	charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Formatter: "{b}: {c}"}),
	// )

	// pie.Render(w)

	// generate heatmap data
	Nx := 10
	My := 10
	array := make([][3]int, Nx*My)
	for n := 0; n < Nx; n++ {
		for m := 0; m < My; m++ {
			array[n*Nx+m] = [3]int{n, m, rand.Int() % 100}

		}
	}
	xlabels := [...]string{"0.0", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7", "0.8", "0.9"}
	ylabels := [...]string{"0.0", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7", "0.8", "0.9"}

	items := make([]opts.HeatMapData, 0)
	for i := 0; i < len(array); i++ {
		items = append(items, opts.HeatMapData{Value: [3]interface{}{array[i][0], array[i][1], array[i][2]}})
	}

	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "basic heatmap example",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type:      "category",
			Data:      xlabels,
			SplitArea: &opts.SplitArea{Show: opts.Bool(true)},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type:      "category",
			Data:      ylabels,
			SplitArea: &opts.SplitArea{Show: opts.Bool(true)},
		}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Calculable: opts.Bool(true),
			Min:        0,
			Max:        100,
			InRange: &opts.VisualMapInRange{
				Color: []string{"#50a3ba", "#eac736", "#d94e5d"},
			},
		}),
	)
	hm.AddSeries("heatmap", items)

	// var buf bytes.Buffer
	snippet := hm.RenderSnippet()
	fmt.Println("snippet = ", snippet)
	fmt.Println("snippet.Element = ", snippet.Element)
	fmt.Println("snippet.Option = ", snippet.Option)
	fmt.Println("snippet.Script = ", snippet.Script)

	w.Write([]byte(snippet.Element))
	w.Write([]byte(snippet.Script))

	// Try rendering the heatmap into a buffer

	// cs_renderer := hm.(chartrender.ChartSnippet)

	// r := c.(chartrender.Renderer)
	// err = r.Render(&buf)

	// generate chart and write it to io.Writer
	// f, _ := os.Create("pie.html")
	// pie.Renderer = newSnippetRenderer(pie, pie.Validate)
	// var htmlSnippet template.HTML = renderToHtml(hm)

	// fmt.Println("htmlSnippet = ", htmlSnippet)

	// type dummy struct {
	// 	Value int
	// }

	// dummy_t := dummy{Value: rand.Int() % 100}
	// tmpl := template.Must(template.ParseFiles("templates/sparta_plot.tmpl"))
	// tmpl.Execute(w, dummy_t)

	return 0, nil
}
