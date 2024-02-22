package main

import (
//      "bytes"
        "flag"
        "fmt"
        "github.com/kumina/openvpn_exporter/exporters"
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        "log"
        "net/http"
        "os/exec"
        "strings"
        "time"
)

var serviceOpenVPN = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "service_openvpn",
        Help: "Indicates whether the OpenVPN service is up (1) or down (0).",
})

func main() {
        var (
                listenAddress      = flag.String("web.listen-address", ":9176", "Address to listen on for web interface and telemetry.")
                metricsPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
                openvpnStatusPaths = flag.String("openvpn.status_paths", "/var/log/openvpn-status.log", "Paths at which OpenVPN places its status files.")
                ignoreIndividuals  = flag.Bool("ignore.individuals", false, "If ignoring metrics for individuals.")
        )
        flag.Parse()

        log.Printf("Starting OpenVPN Exporter\n")
        log.Printf("Listen address: %v\n", *listenAddress)
        log.Printf("Metrics path: %v\n", *metricsPath)
        log.Printf("openvpn.status_path: %v\n", *openvpnStatusPaths)
        log.Printf("Ignore Individuals: %v\n", *ignoreIndividuals)

        exporter, err := exporters.NewOpenVPNExporter(strings.Split(*openvpnStatusPaths, ","), *ignoreIndividuals)
        if err != nil {
                panic(err)
        }
        prometheus.MustRegister(exporter)
        prometheus.MustRegister(serviceOpenVPN)

        go updateOpenVPNServiceStatus()

        http.Handle(*metricsPath, promhttp.Handler())
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte(`
                        <html>
                        <head><title>OpenVPN Exporter</title></head>
                        <body>
                        <h1>OpenVPN Exporter</h1>
                        <p><a href='` + *metricsPath + `'>Metrics</a></p>
                        </body>
                        </html>`))
        })
        log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func updateOpenVPNServiceStatus() {
        for {
                isActive, err := isServiceActive("openvpn")
                if err != nil {
                        log.Printf("Error checking OpenVPN service status: %v", err)
                } else {
                        if isActive {
                                serviceOpenVPN.Set(1)
                        } else {
                                serviceOpenVPN.Set(0)
                        }
                }
                time.Sleep(5 * time.Second)
        }
}

func isServiceActive(serviceName string) (bool, error) {
    cmd := exec.Command("systemctl", "is-active", "--quiet", serviceName)

    // Run the command without capturing output
    if err := cmd.Run(); err != nil {
        // Check if exit status is 3 (unknown) and treat it as inactive
        if exitError, ok := err.(*exec.ExitError); ok {
            if status, ok := exitError.Sys().(syscall.WaitStatus); ok && status.ExitStatus() == 3 {
                return false, nil
            }
        }

        return false, fmt.Errorf("failed to check service status: %v", err)
    }

    return true, nil
}
