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

    socket.on("hello_back", (msg) => {
      console.log('getting from helloback ', msg.got)
    })
    socket.emit("hello", { test: "test" })
    // socket.setTimeout(() => socket.emit("hello", { test: "test" }), 2000);

  });

  sleep(5);

}
