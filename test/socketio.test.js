import { check } from "k6";
import { io } from "k6/x/socketio";
import { sleep } from "k6";

export const options = {
  thresholds: {
    checks: ["rate==1"],
  },
};

export default function () {

  io("http://localhost:4000", {}, (socket) => {
    let connected = false;

    socket.on("connect", (data) => {
      console.log('yo')
      console.log('data', data)
    })

    socket.on("disconnect", () => {
      console.log('closed')
    })

    socket.on("hello_back", (msg) => {
      console.log('getting from helloback ', msg.got)
    })
    socket.emit("hello", { test: "test" })

  });

  sleep(30);

}
