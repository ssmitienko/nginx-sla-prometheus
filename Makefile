nginx-sla-prometheus: main.go metrics.go
	CGO_ENABLED=0 go build -o nginx-sla-prometheus

docker:
	docker build -t nginx-sla-prometheus .

install: nginx-sla-prometheus
	mkdir -p /opt/bin
	install -o root -g root -m 0755 nginx-sla-prometheus /opt/bin/nginx-sla-prometheus
	install -o root -g root -m 0644 nginx-sla-prometheus.service /etc/systemd/system/nginx-sla-prometheus

