package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
)

/*
{
  "http": {
    "services": {
      "test-service": {
        "loadBalancer": {
          "servers": [
            {
              "url": "http://100.92.98.64:3000/"
            }
          ]
        }
      }
    },
    "routers": {
      "test": {
        "rule": "Host(`file-test.shorty.systems`)",
        "tls": {
          "certresolver": "default"
        },
        "service": "test-service"
      }
    }
  }
}
*/

/*
	{
	  "entryPoints": {
	    "tcpep": {
	      "address": ":3179"
	    },
	    "udpep": {
	      "address": ":3179/udp"
	    }
	  }
	}
*/
type TraefikServer struct {
	URL string `json:"url"`
}

type TraefikService struct {
	LoadBalancer struct {
		Servers []TraefikServer `json:"servers"`
	} `json:"loadBalancer"`
}

type TraefikTLS struct {
	CertResolver string `json:"certresolver"`
}

type TraefikRouter struct {
	Rule    string     `json:"rule"`
	TLS     TraefikTLS `json:"tls"`
	Service string     `json:"service"`
}

type HttpService struct {
	URL          string `json:"url"`
	PublicDomain string `json:"publicDomain"`
	Name         string `json:"name"`
}

// type SocketService struct {
// 	Port    string
// 	Address string
// 	Name    string
// }

func loadFromDisk() ([]HttpService, []error) {
	files, err := fs.Glob(os.DirFS("."), "config/*.json")
	if err != nil {
		log.Fatal(err)
	}
	var services []HttpService
	var errs []error
	for _, file := range files {
		// Read the file
		f, err := os.ReadFile(file)
		if err != nil {
			log.Print(err)
			errs = append(errs, err)
		}
		// Parse the file
		var raw []HttpService
		err = json.Unmarshal(f, &raw)
		if err != nil {
			log.Print(err)
			errs = append(errs, err)
		}
		name := path.Base(file)
		name = name[:len(name)-5]
		for _, service := range raw {
			service.Name = name + "-" + service.Name
			services = append(services, service)
		}
	}
	return services, errs
}

func main() {
	defaultTLS := TraefikTLS{
		CertResolver: "default",
	}

	httpServices, _ := loadFromDisk()
	token := os.Getenv("TOKEN")

	// socketServices := []SocketService{
	// 	{
	// 		Port:    "3001",
	// 		Address: "100.92.98.64:3000",
	// 		Name:    "test-socket",
	// 	},
	// }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/traefik" {
			// Write a json response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			out := make(map[string]interface{})
			if len(httpServices) > 0 {
				out["http"] = make(map[string]interface{})
				for _, service := range httpServices {
					s := TraefikService{}
					s.LoadBalancer.Servers = append(s.LoadBalancer.Servers, TraefikServer{URL: service.URL})
					if out["http"].(map[string]interface{})["services"] == nil {
						out["http"].(map[string]interface{})["services"] = make(map[string]interface{})
					}
					out["http"].(map[string]interface{})["services"].(map[string]interface{})["manager-"+service.Name] = s

					r := TraefikRouter{
						Rule:    "Host(`" + service.PublicDomain + "`)",
						TLS:     defaultTLS,
						Service: "manager-" + service.Name,
					}
					if out["http"].(map[string]interface{})["routers"] == nil {
						out["http"].(map[string]interface{})["routers"] = make(map[string]interface{})
					}
					out["http"].(map[string]interface{})["routers"].(map[string]interface{})["manager-"+service.Name] = r
				}
			}
			// // This is not possible as of now. See: https://github.com/traefik/traefik/issues/6551
			// if len(socketServices) > 0 {
			// 	out["entryPoints"] = make(map[string]interface{})
			// 	out["tcp"] = make(map[string]interface{})
			// 	out["udp"] = make(map[string]interface{})
			// 	for _, service := range socketServices {
			// 		out["entryPoints"].(map[string]interface{})["manager-tcp-"+service.Name] = map[string]interface{}{
			// 			"address": ":" + service.Port + "/tcp",
			// 		}
			// 		out["entryPoints"].(map[string]interface{})["manager-udp-"+service.Name] = map[string]interface{}{
			// 			"address": ":" + service.Port + "/udp",
			// 		}
			// 		if out["tcp"].(map[string]interface{})["services"] == nil {
			// 			out["tcp"].(map[string]interface{})["services"] = make(map[string]interface{})
			// 		}
			// 		if out["tcp"].(map[string]interface{})["routers"] == nil {
			// 			out["tcp"].(map[string]interface{})["routers"] = make(map[string]interface{})
			// 		}
			// 		out["tcp"].(map[string]interface{})["services"].(map[string]interface{})["manager-tcp-"+service.Name] = map[string]interface{}{
			// 			"loadBalancer": map[string]interface{}{
			// 				"servers": []map[string]interface{}{
			// 					{
			// 						"address": service.Address,
			// 					},
			// 				},
			// 			},
			// 		}
			// 		out["tcp"].(map[string]interface{})["routers"].(map[string]interface{})["manager-tcp-"+service.Name] = map[string]interface{}{
			// 			"entryPoints": []string{"manager-tcp-" + service.Name},
			// 			"service":    "manager-tcp-" + service.Name,
			// 		}

			// 		if out["udp"].(map[string]interface{})["services"] == nil {
			// 			out["udp"].(map[string]interface{})["services"] = make(map[string]interface{})
			// 		}
			// 		if out["udp"].(map[string]interface{})["routers"] == nil {
			// 			out["udp"].(map[string]interface{})["routers"] = make(map[string]interface{})
			// 		}
			// 		out["udp"].(map[string]interface{})["services"].(map[string]interface{})["manager-udp-"+service.Name] = map[string]interface{}{
			// 			"loadBalancer": map[string]interface{}{
			// 				"servers": []map[string]interface{}{
			// 					{
			// 						"address": service.Address,
			// 					},
			// 				},
			// 			},
			// 		}
			// 		out["udp"].(map[string]interface{})["routers"].(map[string]interface{})["manager-udp-"+service.Name] = map[string]interface{}{
			// 			"entryPoints": []string{"manager-udp-" + service.Name},
			// 			"service":    "manager-udp-" + service.Name,
			// 		}
			// 	}
			// }
			json.NewEncoder(w).Encode(out)
			return

		}
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Hello, World!")
			return
		}
		if r.URL.Path == "/reload" {
			if token != r.URL.Query().Get("token") {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, "Unauthorized")
				return
			}
			newServices, errors := loadFromDisk()
			if len(errors) > 0 && r.URL.Query().Get("force") != "true" {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error loading services")
				return
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Reloaded")
			httpServices = newServices
			return
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Not found")

	})

	loadFromDisk()
	log.Fatal(http.ListenAndServe(":3000", nil))
}
