package configurator

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/tiredsosha/admin/tools/logger"
	"gopkg.in/yaml.v3"
)

// структура конфига
type conf struct {
	Broker   string `yaml:"broker"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Port     int    `yaml:"httpPort"`
	MqttOn   bool   `yaml:"mqttActive"`
	StatusOn bool   `yaml:"statusActive"`
}

type ipConfig struct {
	Zones []ipZone `yaml:"zones"`
}

type ipZone struct {
	ID         string         `yaml:"id"`
	InnerZones []ipInnerZone  `yaml:"innerZones"`
	Status     map[string]int `yaml:"status"`
}

type ipInnerZone struct {
	ID     string         `yaml:"id"`
	Status map[string]int `yaml:"status"`
}

type ipRecord struct {
	Zone      string `yaml:"zone" json:"zone"`
	Equipment string `yaml:"equipment" json:"equipment"`
	IP        string `yaml:"ip" json:"ip"`
}

// поиск конфига на диске
func getConf(file string, cnf any) error {
	yamlFile, err := os.ReadFile(file)
	if err == nil {
		err = yaml.Unmarshal(yamlFile, cnf)
	}

	return err
}

// валидация конфига, валидация проходится если все заполнено
func validateConf(cfg *conf) error {
	var err error
	v := reflect.ValueOf(*cfg)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := typeOfS.Field(i).Name
		value := v.Field(i).Interface()
		if value == "" {
			err = errors.New("config: " + strings.ToLower(field) + " field is emtpy/nonexist")
			break
		}
		err = nil
	}
	return err
}

// создание дефолтного конфига и запись его в файл
func confFile() *conf {
	logger.Warn.Println("making a default config")
	confDef := conf{
		Broker:   "127.0.0.1",
		Username: "admin",
		Password: "password",
		StatusOn: true,
		MqttOn:   false,
		Port:     8080,
	}

	yamlData, _ := yaml.Marshal(confDef)

	if err := os.WriteFile("./configs/config.yaml", yamlData, 0644); err != nil {
		logger.Error.Fatal("can't to write default conf into the file")
	}
	return &confDef
}

func ConfInit() *conf {
	// создаем пустую версию конфига, или ссылку на него
	cfg := &conf{}
	// если у нас нет файла конфиг, то мы создаем дефотный конфиг
	if err := getConf("./configs/config.yaml", cfg); err != nil {
		logger.Error.Println(err)
		cfg = confFile()
	}
	// если у нас есть конфиг и он не проходит валидацию, то вы выходим из приложения
	if err := validateConf(cfg); err != nil {
		logger.Warn.Println(err)
		logger.Error.Fatal("EXITING")
	}

	return cfg
}

func ConfSubInit() {
	// если наше приложение не видит побочных конфигов
	if err := configRelay(); err != nil {
		logger.Warn.Println(err)
		logger.Error.Fatal("EXITING")
	}
	if err := configPC(); err != nil {
		logger.Warn.Println(err)
		logger.Error.Fatal("EXITING")
	}
	if err := configPJ(); err != nil {
		logger.Warn.Println(err)
		logger.Error.Fatal("EXITING")
	}
	if err := configDALI(); err != nil {
		logger.Warn.Println(err)
		logger.Error.Fatal("EXITING")
	}
	if err := GenerateIp(); err != nil {
		logger.Warn.Println(err)
		logger.Warn.Println("ip.yaml не обновлён, используется существующий файл")
	}
}

// GenerateIp один раз создаёт таблицу IP-адресов при запуске приложения.
func GenerateIp() error {
	data, err := os.ReadFile("./configs/status.yaml")
	if err != nil {
		logger.Error.Printf("read status.yaml: %v", err)
		return errors.New("read status.yaml")
	}

	var status ipConfig
	if err := yaml.Unmarshal(data, &status); err != nil {
		logger.Error.Printf("unmarshal status.yaml: %v", err)
		return errors.New("unmarshal status.yaml")
	}

	records := make([]ipRecord, 0)
	addRecords := func(zone string, statuses map[string]int) error {
		equipmentNames := make([]string, 0, len(statuses))
		for equipment := range statuses {
			equipmentNames = append(equipmentNames, equipment)
		}
		sort.Strings(equipmentNames)

		for _, equipment := range equipmentNames {
			ip, err := equipmentIP(zone, equipment)
			if err != nil {
				if equipment == "pc_1" {
					logger.Error.Printf("IP для обязательного оборудования %q в зоне %q не найден", equipment, zone)
					return err
				}
				logger.Warn.Printf("IP для необязательного оборудования %q в зоне %q не найден, запись пропущена", equipment, zone)
				continue
			}
			records = append(records, ipRecord{
				Zone:      zone,
				Equipment: equipment,
				IP:        ip,
			})
		}
		return nil
	}

	for _, zone := range status.Zones {
		if err := addRecords(zone.ID, zone.Status); err != nil {
			return err
		}
		for _, innerZone := range zone.InnerZones {
			if err := addRecords(innerZone.ID, innerZone.Status); err != nil {
				return err
			}
		}
	}
	if len(records) == 0 {
		logger.Error.Println("не удалось создать IP-таблицу: записи отсутствуют")
		return errors.New("IP records are empty")
	}

	output, err := yaml.Marshal(records)
	if err != nil {
		logger.Error.Printf("marshal ip.yaml: %v", err)
		return errors.New("marshal ip.yaml")
	}
	if err := writeIpFileAtomic("./configs/ip.yaml", output, 0644); err != nil {
		logger.Error.Printf("write ip.yaml: %v", err)
		return errors.New("write ip.yaml")
	}

	logger.Info.Printf("ip.yaml generated: %d records", len(records))
	return nil
}

// writeIpFileAtomic заменяет IP-файл только после полной записи нового содержимого.
func writeIpFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmpFile, err := os.CreateTemp(filepath.Dir(path), ".ip.yaml.tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if err := tmpFile.Chmod(perm); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func equipmentIP(zone, equipment string) (string, error) {
	if relays, ok := Relays[zone]; ok {
		if ip, ok := relays[equipment]; ok && ip != "" {
			return ip, nil
		}
	}

	if equipment == "pc_1" {
		if pc, ok := PC[zone]; ok {
			if ip, ok := pc["ip"]; ok && ip != "" {
				return ip, nil
			}
		}
	} else if projectors, ok := PJ[zone]; ok {
		if ip, ok := projectors[equipment]; ok && ip != "" {
			return ip, nil
		}
	}

	return "", errors.New("IP not found")
}
