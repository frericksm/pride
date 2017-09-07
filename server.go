package main

import (
	"fmt"
	"sort"
	"log"
	"net/http"


	"github.com/urfave/cli"
	"os"
	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"

	"github.com/frericksm/pride/bundle"
	"github.com/frericksm/pride/resource"
	"github.com/frericksm/pride/utils"
	"github.com/frericksm/pride/context"
)

var schema *graphql.Schema


//Parst das Schema aus dem Package 'bundle'
func init() {
	var err error
	schema, err = graphql.ParseSchema(bundle.Schema, &bundle.Resolver{})
	if err != nil {
		panic(err)
	}
}

// 
func bundleRootDir(c *cli.Context) string {

	dir := c.GlobalString("dir")
	if dir != "" {
		return dir
	} 
	
	cwd, error := os.Getwd()
	utils.Check(error)
	return cwd
}

func serve(c *cli.Context) error {
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}))

	bundleRootDir := bundleRootDir(c)
	log.Println(fmt.Sprintf("Serving directory: %s", bundleRootDir))
	
	ctxHandler1 := context.Handler{
		BundleRootDir: bundleRootDir,
		Handler: &relay.Handler{
			Schema: schema,
		},
	}
	http.Handle("/query", &ctxHandler1)

	ctxHandler2 := context.Handler{
		BundleRootDir: bundleRootDir,
		Handler: &resource.Handler{},
		}
	http.Handle("/bundles/", &ctxHandler2)
	
	port := fmt.Sprintf(":%d",c.Int("port"))
	log.Println(fmt.Sprintf("Listening on port %d", c.Int("port")))
	log.Fatal(http.ListenAndServe(port, nil))
	return nil
}

func build(c *cli.Context) error {
	log.Println("Building ...")
	return nil
}

func main() {	
	app := cli.NewApp()
	app.Name = "pride"
	app.Version = "0.0.1"
	app.Usage =  `Unterstützt die Entwicklung von Prozessen in der Prozess-IDE 
           der ISP und bietet weitere Funktionen zur Unterstützung des
           Deployments der Prozess-Bundle`

	app.Commands = []cli.Command{
		{
			Name:    "serve",
			Usage:   "Startet einen HTTP-Server", 
			Description:
			`Startet einen HTTP-Server, der das aktuelle oder das über die globale 
   Option 'dir' gesetzte Verzeichnis dem ISP RemoteLocationExplorer verfügbar 
   macht. 
 
   Der ISP-Applicationserver verwendet diesen HTTP-Server als Quelle für 
   Prozessdefinitionen. In der ISP kann eine Prozessdefinitionen, die vom 
   HTTP-Server bereitgestellt wird, im Prozess-Editor bearbeitet und in der 
   Prozess-Engine ausgeführt werden.`,
			Action:  serve,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name: "port, p",
					Value: 8190,
					Usage: `Der ` + "`PORT`" + ` an dem sich der Server bindet. Muß ein Wert 
                         zwischen 8190 bis 9190 sein.`,
				},
			},
		},
		{
			Name:    "build",
			Usage:   "Builds the bundle jar file",
			Action:  build,
		},
	}

	

	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "dir, d",
			Usage: "Das `DIRECTORY` " + `dessen Unterverzeichnisse
	Bundle-Verzeichnisse sind`,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}

var page = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.css" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/query", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}

			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)
