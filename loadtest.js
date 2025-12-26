import http from "k6/http";
import { check } from "k6";

export const options = {
  stages: [
    { duration: "10s", target: 50 },
    { duration: "30s", target: 100 },
    { duration: "10s", target: 0 },
  ],
};

export default function () {
  const payload = JSON.stringify({
    rps: Math.random() * 200,
    cpu: Math.random() * 100,
  });
  
  const res = http.post("http://localhost:8080/metric", payload, {
    headers: { "Content-Type": "application/json" },
  });
  
  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time < 500ms": (r) => r.timings.duration < 500,
  });
}
