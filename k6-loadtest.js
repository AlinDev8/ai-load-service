import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "10s", target: 20 },
    { duration: "20s", target: 50 }, 
    { duration: "30s", target: 50 },
    { duration: "10s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500"],
  },
};

export default function () {
  const payload = JSON.stringify({
    rps: Math.random() * 200,
    cpu: Math.random() * 100,
    timestamp: Date.now(),
  });
  
  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };
  
  const res = http.post("http://localhost:8080/metric", payload, params);
  
  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time < 500ms": (r) => r.timings.duration < 500,
  });
  
  sleep(0.1);
}
