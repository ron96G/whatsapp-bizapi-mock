from locust import HttpUser, between, task
import json

class TestUser(HttpUser):
    wait_time = between(1, 5)
    
    auth_token = ""
    data = {
        "to": "123123123",
        "type": "text",
        "recipient_type": "individual",
        "text": {
            "body": "Hello World!"
        }
    }
    def on_start(self):
        response = self.client.post("/users/login", auth=("admin", "secret"))
        json_response=response.json()
        print(json_response["users"][0]["token"])
        self.auth_token = json_response["users"][0]["token"]

    @task
    def post_messages(self):
        self.client.post("/messages", 
            headers={"authorization": "Bearer " + self.auth_token},
            data=json.dumps(self.data),)
