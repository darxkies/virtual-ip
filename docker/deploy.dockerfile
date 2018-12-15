FROM scratch

COPY virtual-ip /

ENTRYPOINT ["/virtual-ip"]
