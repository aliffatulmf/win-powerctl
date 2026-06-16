package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host     string
	Port     int
	Password string
}

func Load(path string) *Config {
	cfg := &Config{
		Host:     "0.0.0.0",
		Port:     10125,
		Password: "changeme",
	}

	f, err := os.Open(path)
	if err != nil {
		generate(path, cfg)
		return cfg
	}
	defer f.Close()

	var section string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' || line[0] == ';' {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			continue
		}
		if section == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch section + "." + key {
		case "server.host":
			cfg.Host = val
		case "server.port":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.Port = n
			}
		case "auth.password":
			cfg.Password = val
		}
	}

	if err := scanner.Err(); err != nil {
		return cfg
	}
	return cfg
}

func generate(path string, cfg *Config) {
	content := fmt.Sprintf(`[server]
host = %s
port = %d

[auth]
password = %s
`, cfg.Host, cfg.Port, cfg.Password)

	os.WriteFile(path, []byte(content), 0644)
	fmt.Printf("Generated default config: %s\n", path)
}
