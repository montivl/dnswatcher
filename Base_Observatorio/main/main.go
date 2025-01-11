// Este archivo implementa la lógica principal del programa para recolectar y analizar datos DNS.
// Utiliza el software Observatorio para tareas de recopilación, análisis y utilidades GeoIP.

package main

import (
	"fmt"
	"os"

	"github.com/niclabs/Observatorio/dataAnalyzer"
	"github.com/niclabs/Observatorio/dataCollector"
	"github.com/niclabs/Observatorio/geoIPUtils"
	"gopkg.in/yaml.v2"
)

// Config estructura la configuración cargada desde un archivo YAML.
// Define parámetros de ejecución, configuración de base de datos y ajustes de GeoIP.
type Config struct {
	RunArguments struct {
		InputFilepath     string   `yaml:"inputfilepath"`
		DontProbeFilepath string   `yaml:"dontprobefilepath"`
		Verbose           bool     `yaml:"verbose"`
		Concurrency       int      `yaml:"concurrency"`
		DropDatabase      bool     `yaml:"dropdatabase"`
		Debug             bool     `yaml:"debug"`
		DnsServers        []string `yaml:"dnsservers"`
	} `yaml:"runargs"`
	Database struct {
		DatabaseName string `yaml:"dbname"`
		Username     string `yaml:"dbuser"`
		Password     string `yaml:"dbpass"`
		Host         string `yaml:"dbhost"`
		Port         int    `yaml:"dbport"`
	} `yaml:"database"`
	Geoip struct {
		GeoipPath            string `yaml:"geoippath"`
		GeoipAsnFilename     string `yaml:"geoipasnfilename"`
		GeoipCountryFilename string `yaml:"geoipcountryfilename"`
		GeoipLicenseKey      string `yaml:"geoiplicensekey"`
	} `yaml:"geoip"`
}

// CONFIG_FILE define la ubicación predeterminada del archivo de configuración.
var CONFIG_FILE = "config.yml"

// main es el punto de entrada principal del programa.
// Responsable de cargar la configuración, inicializar recursos y coordinar las tareas de recolección y análisis de datos.
func main() {

	// 1. Cargar archivo de configuración.
	f, err := os.Open(CONFIG_FILE)
	if err != nil {
		fmt.Printf("Can't open configuration file: " + err.Error())
		return
	}
	defer f.Close()
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Printf("Can't decode configuration: " + err.Error())
		return
	}
	// 2. Verificar servidores DNS configurados.
	if len(cfg.RunArguments.DnsServers) == 0 {
		fmt.Printf("you must add at least one dns server in the config file.")
		return
	}

	// 3. Inicializar bases de datos GeoIP.
	var geoipDB = geoIPUtils.InitGeoIP(cfg.Geoip.GeoipPath, cfg.Geoip.GeoipCountryFilename, cfg.Geoip.GeoipAsnFilename, cfg.Geoip.GeoipLicenseKey)

	// 4. Inicializar módulo de recolección de datos.
	err = dataCollector.InitCollect(cfg.RunArguments.DontProbeFilepath, cfg.RunArguments.DropDatabase, cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.DatabaseName, geoipDB, cfg.RunArguments.DnsServers)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 5. Comenzar la recolección de datos.
	runId := dataCollector.StartCollect(cfg.RunArguments.InputFilepath, cfg.RunArguments.Concurrency, cfg.Database.DatabaseName, cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.RunArguments.Debug, cfg.RunArguments.Verbose)

	// 6. Cerrar recursos GeoIP.
	geoIPUtils.CloseGeoIP(geoipDB)

	// 7. Analizar datos recolectados.
	fmt.Println("Analyzing Data...")
	dataAnalyzer.AnalyzeData(runId, cfg.Database.DatabaseName, cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port)

}
