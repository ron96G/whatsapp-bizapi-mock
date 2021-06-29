import http from 'k6/http';
import encoding from 'k6/encoding';
import { check } from 'k6';

const username = 'admin';
const password = 'topsecret123!';
const baseUrl = 'https://localhost:9090/v1'

export default function () {
  const credentials = `${username}:${password}`;
  const encodedCredentials = encoding.b64encode(credentials);

  var params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Basic ${encodedCredentials}`
    },
  };
  let res = http.post(baseUrl+'/users/login', {},params);

  check(res, {
    'status is 200': (r) => r.status === 200
  });

  let bearerToken = res.json().users[0].token

  let payload = JSON.stringify({
    "to": "49170123123123",
    "type": "text",
    "recipient_type": "individual",
    "text": {
      "body": "This is a sample text!"
    }
  });
  params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${bearerToken}`
    },
  };

  res = http.post(baseUrl+'/messages', payload, params);
}
