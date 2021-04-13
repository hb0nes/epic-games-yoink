FROM golang:buster

RUN apt update

RUN apt install -y scrot xvfb git wget

RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb

RUN apt install -y ./google-chrome-stable_current_amd64.deb

ADD https://api.github.com/repositories/321185660/git/refs/heads/main version.json

RUN git clone https://www.github.com/hb0nes/free-game-snatcher-epicstore.git yoink

WORKDIR yoink

RUN go get -u

RUN echo "Current directory: ${PWD}"

CMD Xvfb :99 -screen 0 1024x768x16 & DISPLAY=:99 go run *.go
