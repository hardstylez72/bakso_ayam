FROM alpine
WORKDIR "/opt"

ADD ./bakso_ayam bakso_ayam
RUN chmod +x /opt/bakso_ayam

ENTRYPOINT ["/opt/bakso_ayam"]