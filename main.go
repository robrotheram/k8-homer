package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type HomerConfig struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Logo     string `yaml:"logo"`
	Header   bool   `yaml:"header"`
	Footer   string `yaml:"footer"`
	Theme    string `yaml:"theme"`
	Colors   struct {
		Light struct {
			HighlightPrimary   string `yaml:"highlight-primary"`
			HighlightSecondary string `yaml:"highlight-secondary"`
			HighlightHover     string `yaml:"highlight-hover"`
			Background         string `yaml:"background"`
			CardBackground     string `yaml:"card-background"`
			Text               string `yaml:"text"`
			TextHeader         string `yaml:"text-header"`
			TextTitle          string `yaml:"text-title"`
			TextSubtitle       string `yaml:"text-subtitle"`
			CardShadow         string `yaml:"card-shadow"`
			Link               string `yaml:"link"`
			LinkHover          string `yaml:"link-hover"`
		} `yaml:"light"`
		Dark struct {
			HighlightPrimary   string `yaml:"highlight-primary"`
			HighlightSecondary string `yaml:"highlight-secondary"`
			HighlightHover     string `yaml:"highlight-hover"`
			Background         string `yaml:"background"`
			CardBackground     string `yaml:"card-background"`
			Text               string `yaml:"text"`
			TextHeader         string `yaml:"text-header"`
			TextTitle          string `yaml:"text-title"`
			TextSubtitle       string `yaml:"text-subtitle"`
			CardShadow         string `yaml:"card-shadow"`
			Link               string `yaml:"link"`
			LinkHover          string `yaml:"link-hover"`
		} `yaml:"dark"`
	} `yaml:"colors"`
	Message struct {
		Style   string `yaml:"style"`
		Title   string `yaml:"title"`
		Icon    string `yaml:"icon"`
		Content string `yaml:"content"`
	} `yaml:"message"`
	Links []struct {
		Name   string `yaml:"name"`
		Icon   string `yaml:"icon"`
		URL    string `yaml:"url"`
		Target string `yaml:"target,omitempty"`
	} `yaml:"links"`
	Services []HomerService `yaml:"services"`
}

type HomerService struct {
	Name  string             `yaml:"name"`
	Icon  string             `yaml:"icon"`
	Items []HomerServiceItem `yaml:"items"`
}

type HomerServiceItem struct {
	Name     string `yaml:"name"`
	Logo     string `yaml:"logo"`
	Subtitle string `yaml:"subtitle"`
	Tag      string `yaml:"tag"`
	Keywords string `yaml:"keywords,omitempty"`
	URL      string `yaml:"url"`
	Target   string `yaml:"target,omitempty"`
}

//go:embed default.yaml
var defaultConfigFS embed.FS

func overwriteStructValues(s1, s2 reflect.Value) {
	// Get the type of the struct
	sType := s1.Type()
	// Iterate over fields of the struct
	for i := 0; i < sType.NumField(); i++ {
		fieldName := sType.Field(i).Name
		// Get the field value of the second struct
		fieldValue := s2.FieldByName(fieldName)
		// Check if field value is not empty
		if fieldValue.Kind() != reflect.Invalid && !isEmptyValue(fieldValue) {
			// If the field is a struct, recursively overwrite its fields
			if fieldValue.Kind() == reflect.Struct {
				overwriteStructValues(s1.Field(i), fieldValue)
			} else {
				// Set the field value of the first struct
				s1.Field(i).Set(fieldValue)
			}
		}
	}
}

// Function to check if a value is empty
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	default:
		return false
	}
}

func newConfig() HomerConfig {
	yfile, err := defaultConfigFS.ReadFile("default.yaml")
	if err != nil {
		log.Fatal(err)
	}
	var config HomerConfig
	yaml.Unmarshal(yfile, &config)
	return config
}

func (config *HomerConfig) updateFromFile(files string) error {
	yfile, err := os.ReadFile(files)
	if err != nil {
		return err
	}
	var cfg HomerConfig
	yaml.Unmarshal(yfile, &cfg)

	s1Value := reflect.ValueOf(config).Elem()
	s2Value := reflect.ValueOf(&cfg).Elem()
	overwriteStructValues(s1Value, s2Value)
	return nil
}

func k8LocalClinet() (*rest.Config, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user home dir: %v", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to find kubernetes config: %v", err)
	}
	return kubeConfig, nil
}

func getClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = k8LocalClinet()
		if err != nil {
			fmt.Printf("error getting user home dir: %v\n", err)
			os.Exit(1)
		}
	}
	return kubernetes.NewForConfig(config)
}

func (service *HomerService) addItem(newItem HomerServiceItem) {
	for i := range service.Items {
		item := &service.Items[i]
		if item.Name == newItem.Name {
			service.Items[i] = newItem
			return
		}
	}
	service.Items = append(service.Items, newItem)
}

func (cfg *HomerConfig) addServiceItem(serviceName string, newItem HomerServiceItem) {
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		if svc.Name == serviceName {
			svc.addItem(newItem)
			return
		}
	}
}

func (config *HomerConfig) updateServicesFromK8() error {
	client, err := getClient()
	if err != nil {
		return err
	}

	ingresses, err := client.NetworkingV1().Ingresses("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, ingress := range ingresses.Items {
		item := HomerServiceItem{
			Tag:    ingress.Namespace,
			Target: "_blank",
		}

		var serviceName string
		for annotatoion, value := range ingress.Annotations {
			switch annotatoion {
			case "homer.item.name":
				item.Name = value
			case "homer.item.subtitle":
				item.Subtitle = value
			case "homer.item.logo":
				item.Logo = value
			case "homer.service.name":
				serviceName = value
			}
		}

		for _, rule := range ingress.Spec.Rules {
			item.URL = fmt.Sprintf("https://%s", rule.Host)
			break
		}

		config.addServiceItem(serviceName, item)
	}
	return nil
}

func (config *HomerConfig) write() {
	data, err := yaml.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	// os.MkdirAll("assets", 0755)
	os.WriteFile("www/assets/config.yml", data, 0644)
}

func updateConfig() {
	ticker := time.NewTicker(10 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				log.Println("updating config")
				config := newConfig()
				config.updateFromFile("template.yml")
				config.updateServicesFromK8()
				config.write()
			}
		}
	}()
}

func main() {
	updateConfig()

	router := mux.NewRouter()
	router.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.ServeFile(w, r, "www/index.html")
	}).Methods(http.MethodGet, http.MethodHead)
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("www"))))
	log.Println("starting server :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
