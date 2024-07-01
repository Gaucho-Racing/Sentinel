export interface ClientApplication {
  id: string;
  user_id: string;
  secret: string;
  name: string;
  redirect_uris: string[];
  updated_at: Date;
  created_at: Date;
}

export const initClientApplication: ClientApplication = {
  id: "",
  user_id: "",
  secret: "",
  name: "",
  redirect_uris: [],
  updated_at: new Date(),
  created_at: new Date(),
};

/*
{
    "id": "eqHTFAzg1vro",
    "user_id": "348220961155448833",
    "secret": "xaPME4afDUM0QOjkNfPq0Xw5ukm4lbdm",
    "name": "Test Application",
    "redirect_uris": [
      "http://localhost:8080/auth/login"
    ],
    "updated_at": "2024-06-30T20:40:28.06707-07:00",
    "created_at": "2024-06-30T20:39:31.836686-07:00"
  }
*/
