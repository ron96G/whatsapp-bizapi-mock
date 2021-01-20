# WhatsApp Business API Mockserver

This repository contains the mockserver for the WhatsApp Business API (WABiz). 
It is used to perform integration-, system- and load-testing.


## Supported Endpoints

| Endpoint| Usage | Supported |
| :--------------- | :------------- | :------ |
| POST /v1/generate| generate webhook requests| ✅ |
| POST /v1/generate/cancel  | stop generation of webhook requests| ✅ |
| POST /v1/messages| send messages| ✅ |
| POST /v1/users| create user| ✅ |
| DEL /v1/users/{name}| delete user| ✅ |
| POST /v1/users/login| login user| ✅ |
| POST /v1/users/logout| logout user| ✅ |
| POST /v1/media| save media file| ✅ |
| GET /v1/media/{id}| delete media file| ✅ |
| DEL /v1/{media/id}| get media file| ✅ |
| GET /v1/contacts| get contact info for wa_id| ⬜️ |
| XXX /v1/settings/**| setup application settings| ⬜️ |
| XXX /v1/profile/**| setup all profile settings| ⬜️ |
| XXX /v1/stickerpacks/**| all stickerpacks functionality | ⬜️ |
| XXX /v1/certificates/**| all certificates functionality | ⬜️ |
| XXX /v1/account | registration functionality | ⬜️ |
| XXX /v1/account/verify | registration functionality | ⬜️ |

## Functionaliy
The following list shows the core functionality that is currently supported.

1. Generate inbound traffic with different messages and media
2. Generate stati for outbound traffic
3. (TBD) Validate outbound traffic 
4. Rate limiting


## Supported Messages
The following message types are currently supported.
Inbound types are generated and sent via the webhook.
Outbound types are accepted by the messages resource and validated.

| Type | Inbound | Outbound |
| :--- | :---| :--- |
| Text | ✅ | ✅ |
| Image | ✅ | ✅ |
| Audio | ✅ | ✅ |
| Video | ✅ | ✅ |
| Document | ✅ | ✅ |
| Voice | ⬜️ | ⬜️ |
| Location | ⬜️ | ✅ |
| Contact | ⬜️ | ⬜️ |
| Hsm | ⬜️ | ⬜️ |
| System | ⬜️ | ⬜️ |