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

func sparta_title_str(s sparta_info) string {

	t_frame := float64(s.Current_frame) * (s.T_end - s.T_start) / float64(s.Num_frames)
	title_str := fmt.Sprintf("Shot %d frame %d t=%f", s.Shotnr, s.Current_frame, t_frame)

	return title_str
}

// Fetches data and renders it into a heatmap
// Updates the user state: sparta_info.Current_frames
// Return JS code + data for the plot
func fetch_sparta_plot(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	new_state := user_state{}

	// Parse the form and find the requested frame
	err := r.ParseForm()
	if err != nil { // Return a Bad Request if we can't parse the form
		fmt.Fprintf(os.Stdout, "/fetch_sparta_plot: Unable to parse %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return 400, nil
	}
	fmt.Println("/fetch_sparta_plot: range= ", r.Form["range"][0])

	// Get the current frame from the request
	var current_frame int = 0
	if key, ok := r.Form["range"]; ok {
		val, err := strconv.ParseInt(key[0], 10, 32)
		if err != nil {
			return 0, err
		}
		current_frame = int(val)
	}
	fmt.Println("/fetch_sparta_plot:Fetching data for frame ", current_frame)

	title_str := ""

	// If the user is logged in, update the frame the user is on
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("/fetch_sparta_plot: Session token not set")
		w.Write([]byte("You need to log in before browsing shots."))
		return 0, nil

	} else {
		username := a.session_to_user[c.Value]
		fmt.Println("/fetch_sparta_plot: username = ", username)
		// Update state
		new_state.Current_session_id = a.all_user_state[username].Current_session_id
		new_state.Username = username
		new_state.Sparta_state = sparta_info{Shotnr: a.all_user_state[username].Sparta_state.Shotnr,
			T_start:       a.all_user_state[username].Sparta_state.T_start,
			T_end:         a.all_user_state[username].Sparta_state.T_end,
			Num_frames:    a.all_user_state[username].Sparta_state.Num_frames,
			Current_frame: current_frame}

		a.all_user_state[username] = new_state
		fmt.Println("/fetch_sparta_plot: New state: ", a.all_user_state[username])
		title_str = sparta_title_str(a.all_user_state[username].Sparta_state)
	}

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
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "600px",
			Height: "600px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title: title_str,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			// Type:      "value",
			Data:      xlabels,
			SplitArea: &opts.SplitArea{Show: opts.Bool(true)},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			// Type:      "value",
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

	snippet := hm.RenderSnippet()

	// Write the chart into the response
	w.Write([]byte(snippet.Element))
	w.Write([]byte(snippet.Script))

	return 0, nil
}
