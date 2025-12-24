import { check } from "k6";
import { greeting } from "k6/x/socket.io";

export const options = {
  thresholds: {
    checks: ["rate==1"],
  },
};

export default function () {
  check(greeting(""), {
    "greeting()": (str) => str == "Hello, World!",
  });

  check(greeting("Joe"), {
    "greeting('Joe')": (str) => str == "Hello, Joe!",
  });
}
