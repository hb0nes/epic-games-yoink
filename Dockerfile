FROM golang:buster

RUN apt update

RUN apt install -y scrot xvfb git wget

RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb

RUN apt install -y ./google-chrome-stable_current_amd64.deb

ADD https://showcase.api.linx.twenty57.net/UnixTime/tounixtimestamp?datetime=now /tmp/bustcache

RUN git clone --depth 1 https://www.github.com/hb0nes/free-game-snatcher-epicstore.git yoink

WORKDIR yoink

RUN go get -u

RUN echo "Current directory: ${PWD}"

CMD Xvfb :99 -screen 0 1024x768x24 & DISPLAY=:99 go run *.go
