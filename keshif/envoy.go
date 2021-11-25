package keshif

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type SocketAddress struct {
	Address   string `yaml:"address"`
	PortValue int    `yaml:"port_value"`
}

type Address struct {
	SocketAddress SocketAddress `yaml:"socket_address"`
}

type Admin struct {
	Address Address `yaml:"address"`
}

type ExampleListenerName struct {
	ConnectionLimit int `yaml:"connection_limit"`
}

type Listener1 struct {
	ExampleListenerName ExampleListenerName `yaml:"example_listener_name"`
}
type ResourceLimits struct {
	Listener Listener1 `yaml:"listener"`
}

type Envoy struct {
	ResourceLimits ResourceLimits `yaml:"resource_limits"`
}

type StaticLayer struct {
	Envoy Envoy `yaml:"envoy"`
}

type Layer struct {
	Name        string      `yaml:"name"`
	StaticLayer StaticLayer `yaml:"static_layer"`
}

type LayeredRuntime struct {
	Layers []Layer `yaml:"layers"`
}

type Match struct {
	Prefix string `yaml:"prefix"`
}

type EnvoyRoute struct {
	Cluster string `yaml:"cluster"`
}

type Routes struct {
	Match Match      `yaml:"match"`
	Route EnvoyRoute `yaml:"route"`
}

type VirtualHosts struct {
	Name    string   `yaml:"name"`
	Domains []string `yaml:"domains"`
	Routes  []Routes `yaml:"routes"`
}

type RouteConfig struct {
	Name         string         `yaml:"name"`
	VirtualHosts []VirtualHosts `yaml:"virtual_hosts"`
}

type HTTPFilters struct {
	Name string `yaml:"name"`
}

type TypedConfig struct {
	Type        string        `yaml:"@type"`
	CodecType   string        `yaml:"codec_type"`
	StatPrefix  string        `yaml:"stat_prefix"`
	RouteConfig RouteConfig   `yaml:"route_config"`
	HTTPFilters []HTTPFilters `yaml:"http_filters"`
}

type Filter struct {
	Name        string      `yaml:"name"`
	TypedConfig TypedConfig `yaml:"typed_config"`
}

type FilterChains struct {
	Filters []Filter `yaml:"filters"`
}

type Listener struct {
	Address      Address        `yaml:"address"`
	FilterChains []FilterChains `yaml:"filter_chains"`
}

type Endpoint struct {
	Address Address `yaml:"address"`
}

type LbEndpoints struct {
	Endpoint Endpoint `yaml:"endpoint"`
}

type Endpoints struct {
	LbEndpoints []LbEndpoints `yaml:"lb_endpoints"`
}

type LoadAssignment struct {
	ClusterName string      `yaml:"cluster_name"`
	Endpoints   []Endpoints `yaml:"endpoints"`
}

type Cluster struct {
	Name           string         `yaml:"name"`
	Type           string         `yaml:"type"`
	LbPolicy       string         `yaml:"lb_policy"`
	LoadAssignment LoadAssignment `yaml:"load_assignment"`
}

type StaticResources struct {
	Listeners []Listener `yaml:"listeners"`
	Clusters  []Cluster  `yaml:"clusters"`
}

type EnvoyConfig struct {
	StaticResources StaticResources `yaml:"static_resources"`
	Admin           Admin           `yaml:"admin"`
	LayeredRuntime  LayeredRuntime  `yaml:"layered_runtime"`
}

func GetEnvoyConfig(routes map[string]Route) ([]byte, error) {
	config := EnvoyConfig{
		Admin: Admin{
			Address: Address{
				SocketAddress: SocketAddress{
					Address:   "0.0.0.0",
					PortValue: 8001,
				},
			},
		},
		LayeredRuntime: LayeredRuntime{
			Layers: []Layer{{
				Name: "static_layer_0",
				StaticLayer: StaticLayer{
					Envoy: Envoy{
						ResourceLimits: ResourceLimits{
							Listener: Listener1{
								ExampleListenerName: ExampleListenerName{
									ConnectionLimit: 10000,
								},
							},
						},
					},
				},
			}},
		},
	}

	virtualHosts := []VirtualHosts{}
	clusters := []Cluster{}

	for name, route := range routes {
		port, _ := route.Port.Int64()
		newCluster := Cluster{
			Name:     name,
			Type:     "STRICT_DNS",
			LbPolicy: "ROUND_ROBIN",
			LoadAssignment: LoadAssignment{
				ClusterName: name,
				Endpoints: []Endpoints{{
					LbEndpoints: []LbEndpoints{{
						Endpoint: Endpoint{
							Address: Address{
								SocketAddress: SocketAddress{
									Address:   route.Ip,
									PortValue: int(port),
								},
							},
						},
					}},
				}},
			},
		}

		newVirtualHost := VirtualHosts{
			Name:    name,
			Domains: []string{route.Vhost},
			Routes: []Routes{{
				Match: Match{
					Prefix: "",
				},
				Route: EnvoyRoute{
					Cluster: name,
				},
			}},
		}

		clusters = append(clusters, newCluster)
		virtualHosts = append(virtualHosts, newVirtualHost)
	}

	staticResources := StaticResources{}
	staticResources.Clusters = clusters
	staticResources.Listeners = []Listener{{
		Address: Address{
			SocketAddress: SocketAddress{
				Address:   "0.0.0.0",
				PortValue: 80,
			},
		},
		FilterChains: []FilterChains{{
			Filters: []Filter{{
				Name: "envoy.filters.network.http_connection_manager",
				TypedConfig: TypedConfig{
					Type:       "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
					CodecType:  "AUTO",
					StatPrefix: "ingress_http",
					RouteConfig: RouteConfig{
						Name:         "local_route",
						VirtualHosts: virtualHosts,
					},
					HTTPFilters: []HTTPFilters{{
						Name: "envoy.filters.http.router",
					}},
				},
			}},
		}},
	}}

	config.StaticResources = staticResources

	generated, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return generated, nil
}

func createConfigFile(config []byte) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	target := dirname + "/.keshif/envoy.yaml"

	f, err := os.Create(target)
	defer f.Close()

	if err != nil {
		panic(fmt.Sprintf("An error occured while creating envoy configuration: ", err))
	}

	f.Write(config)
}

func GenerateEnvoyConfig(routes map[string]Route) {
	generatedConfig, err := GetEnvoyConfig(routes)

	if err != nil {
		panic(fmt.Sprintf("An error occured while generating envoy configuration: ", err))
	}

	createConfigFile(generatedConfig)
}
