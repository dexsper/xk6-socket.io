import { b32encode } from "k6/x/socket.io";

export default function () {
  console.log(b32encode("Hello, World!"))
}
