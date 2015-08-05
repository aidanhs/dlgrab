FROM aidanhs/chmod

ADD https://github.com/aidanhs/dlgrab/releases/download/0.2/dlgrab-linux-x64 /dlgrab
RUN ["/chmod", "+x", "/dlgrab"]
ENTRYPOINT ["/dlgrab", "--outdir=/out"]
