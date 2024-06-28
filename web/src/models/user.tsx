export interface User {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone_number: string;
  graduate_level: string;
  graduation_year: number;
  major: string;
  shirt_size: string;
  jacket_size: string;
  sae_registration_number: string;
  avatar_url: string;
  verified: boolean;
  subteams: Subteam[];
  roles: string[];
  updated_at: string;
  created_at: string;
}

export interface Subteam {
  id: string;
  name: string;
  created_at: string;
}

export const initUser = {
  id: "",
  first_name: "",
  last_name: "",
  email: "",
  phone_number: "",
  graduate_level: "",
  graduation_year: 0,
  major: "",
  shirt_size: "",
  jacket_size: "",
  sae_registration_number: "",
  avatar_url: "",
  verified: false,
  subteams: [],
  roles: [],
  updated_at: "",
  created_at: "",
};

/*
{
  "id": "348220961155448833",
  "username": "bk1031",
  "first_name": "Bharat",
  "last_name": "Kathi",
  "email": "bkathi@ucsb.edu",
  "phone_number": "",
  "graduate_level": "",
  "graduation_year": 0,
  "major": "",
  "shirt_size": "",
  "jacket_size": "",
  "sae_registration_number": "",
  "avatar_url": "https://cdn.discordapp.com/avatars/348220961155448833/1bedb626ddb6b5a712ee3b172e442eff.png?size=256",
  "verified": false,
  "subteams": [
    {
      "id": "761116347865890816",
      "name": "Electronics",
      "created_at": "2024-06-27T18:23:22.099813-07:00"
    },
    {
      "id": "761331962563919874",
      "name": "Business",
      "created_at": "2024-06-27T18:23:22.104944-07:00"
    },
    {
      "id": "1254572624307290202",
      "name": "Data",
      "created_at": "2024-06-27T18:23:22.11023-07:00"
    }
  ],
  "roles": [
    "d_lead",
    "d_admin",
    "d_officer",
    "d_member",
    "github_BK1031"
  ],
  "updated_at": "2024-06-27T00:34:12.749266-07:00",
  "created_at": "2024-06-27T00:34:12.771085-07:00"
}
*/
