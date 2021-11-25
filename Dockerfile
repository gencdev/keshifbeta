FROM envoyproxy/envoy-dev:663cd1789b3676ed1d5293f21524cb123424027d

COPY envoy.yaml /etc/envoy/envoy.yaml
RUN chmod go+r /etc/envoy/envoy.yaml


