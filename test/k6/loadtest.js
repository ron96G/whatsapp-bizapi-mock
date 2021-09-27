import http from 'k6/http';
import encoding from 'k6/encoding';
import { check } from 'k6';

const username = 'admin';
const password = 'secret';
const new_password = 'secret_new_pwd123!'
const baseUrl = 'https://localhost:9090/v1'
const credentials = `${username}:${password}`;
const encodedCredentials = encoding.b64encode(credentials);
var params = {
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Basic ${encodedCredentials}`
  },
};


export let options = {
  vus: 300,
  duration: '10s',
};


export function setup() {
  let res = http.post(baseUrl+'/users/login', JSON.stringify({"new_password": new_password}), params);

  if (res.status !== 200) {
    params.headers.Authorization = encoding.b64encode(`${username}:${new_password}`)
    res = http.post(baseUrl+'/users/login', "{}", params);
  }

  check(res, {
    'status is 200': (r) => r.status === 200
  });

  return {bearerToken : res.json().users[0].token}
}

export default function (data) {
  let payload = JSON.stringify({
    "to": "49170123123123",
    "type": "text",
    "recipient_type": "individual",
    "text": {
      "body": "This is a sample text!"
    }
  });
  let params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.bearerToken}`
    },
  };

  let res = http.post(baseUrl+'/messages', payload, params);
  check(res, {
    'status is 200': (r) => r.status === 200
  });
}
