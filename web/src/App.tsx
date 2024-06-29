import React from "react";
import axios from "axios";
import {
  DISCORD_CLIENT_ID,
  DISCORD_OAUTH_BASE_URL,
  GITHUB_ORG_URL,
  SENTINEL_API_URL,
  SHARED_DRIVE_URL,
  currentUser,
} from "@/consts/config";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { useNavigate } from "react-router-dom";
import { Separator } from "@/components/ui/separator";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faCheckCircle,
  faLock,
  faPerson,
  faUser,
} from "@fortawesome/free-solid-svg-icons";
import { checkCredentials } from "@/lib/auth";
import Footer from "@/components/Footer";
import Header from "@/components/Header";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import {
  faGithub,
  faGithubAlt,
  faGoogleDrive,
} from "@fortawesome/free-brands-svg-icons";

function App() {
  const navigate = useNavigate();

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);

  const [loginLoading, setLoginLoading] = React.useState(false);
  const [loginAccess, setLoginAccess] = React.useState({});

  const [driveLoading, setDriveLoading] = React.useState(false);
  const [driveAccess, setDriveAccess] = React.useState({});

  const [githubLoading, setGithubLoading] = React.useState(false);
  const [githubAccess, setGithubAccess] = React.useState({});

  React.useEffect(() => {
    checkAuth().then(() => {
      checkLoginAccess();
      checkDriveAccess();
      checkGithubAccess();
    });
  }, []);

  const checkAuth = async () => {
    setAuthCheckLoading(true);
    const currentRoute = window.location.pathname + window.location.search;
    const status = await checkCredentials();
    if (status != 0) {
      navigate(`/auth/login?route=${encodeURIComponent(currentRoute)}`);
    } else {
      setAuthCheckLoading(false);
    }
  };

  const checkLoginAccess = async () => {
    setLoginLoading(true);
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/users/${currentUser.id}/auth`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      setLoginAccess(response.data);
    } catch (error: any) {
      if (!getAxiosErrorMessage(error).includes("No authentication found")) {
        toast(getAxiosErrorMessage(error));
      }
    }
    setLoginLoading(false);
  };

  const registerPassword = async (password: string) => {
    setLoginLoading(true);
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/auth/register`,
        {
          email: currentUser.email,
          password: password,
        },
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      if (response.status == 200) {
        localStorage.setItem("id", response.data.id);
        localStorage.setItem("token", response.data.token);
        checkCredentials();
      }
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
    checkLoginAccess();
  };

  const checkDriveAccess = async () => {
    setDriveLoading(true);
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/users/${currentUser.id}/drive`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      setDriveAccess(response.data);
    } catch (error: any) {
      if (!getAxiosErrorMessage(error).includes("No permissions found")) {
        toast(getAxiosErrorMessage(error));
      }
    }
    setDriveLoading(false);
  };

  const addUserToDrive = async () => {
    setDriveLoading(true);
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/users/${currentUser.id}/drive`,
        {},
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      setDriveAccess(response.data);
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
    checkDriveAccess();
  };

  const removeUserFromDrive = async () => {
    setDriveLoading(true);
    try {
      const response = await axios.delete(
        `${SENTINEL_API_URL}/users/${currentUser.id}/drive`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      setDriveAccess(response.data);
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
    checkDriveAccess();
  };

  const checkGithubAccess = async () => {
    setGithubLoading(true);
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/users/${currentUser.id}/github`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      setGithubAccess(response.data);
    } catch (error: any) {
      if (!getAxiosErrorMessage(error).includes("user does not have")) {
        toast(getAxiosErrorMessage(error));
      }
    }
    setGithubLoading(false);
  };

  const addUserToGithub = async (username: string) => {
    setGithubLoading(true);
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/users/${currentUser.id}/github`,
        {
          username: username,
        },
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      setGithubAccess(response.data);
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
    checkGithubAccess();
  };

  const AuthLoading = () => {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <Card className="border-none p-8" style={{ width: 500 }}>
          <div className="flex flex-col items-center justify-center">
            <img
              src="/logo/mechanic-logo.png"
              alt="Gaucho Racing"
              className="mx-auto h-24"
            />
            <Loader2 className="mt-8 h-16 w-16 animate-spin" />
          </div>
        </Card>
      </div>
    );
  };

  const ProfileField = (props: { label: string; value: string }) => {
    return (
      <div className="mx-2 mt-2 flex">
        <div className="mr-2 font-semibold">{props.label}:</div>
        <div className="text-gray-400">
          {props.value != "" ? props.value : "Not set"}
        </div>
      </div>
    );
  };

  const ProfileCard = () => {
    return (
      <Card className="mr-4 mt-4 w-[500px] p-4">
        <div className="flex items-center justify-start">
          <FontAwesomeIcon icon={faUser} className="h-5 w-5" />
          <h3 className="ml-4">Profile</h3>
        </div>
        <Separator className="my-2" />
        <div className="flex items-center justify-start">
          <Avatar className="mr-4">
            <AvatarImage src={currentUser.avatar_url} />
            <AvatarFallback>CN</AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <p>
              {currentUser.first_name} {currentUser.last_name}
            </p>
            <p className="text-gray-400">{currentUser.email}</p>
          </div>
        </div>
        <ProfileField label="ID" value={currentUser.id} />
        <ProfileField label="Phone Number" value={currentUser.phone_number} />
        <ProfileField
          label="Graduate Level"
          value={currentUser.graduate_level}
        />
        <ProfileField
          label="Graduation Year"
          value={currentUser.graduation_year.toString()}
        />
        <ProfileField label="Major" value={currentUser.major} />
        <ProfileField label="Shirt Size" value={currentUser.shirt_size} />
        <ProfileField label="Jacket Size" value={currentUser.jacket_size} />
        <ProfileField
          label="SAE Member Number"
          value={currentUser.sae_registration_number}
        />
        <ProfileField
          label="Subteams"
          value={currentUser.subteams.map((subteam) => subteam.name).join(", ")}
        />
        <div className="mx-2 mt-2 flex">
          <div className="mr-2 font-semibold">Roles:</div>
          <div className="flex flex-wrap">
            {currentUser.roles.map((role) => (
              <div key={role} className="mx-1 mb-2">
                <Card className="rounded-sm px-1 text-gray-400">
                  <code className="">{role}</code>
                </Card>
              </div>
            ))}
          </div>
        </div>
        <ProfileField
          label="Updated At"
          value={new Date(currentUser.updated_at).toLocaleString()}
        />
        <ProfileField
          label="Created At"
          value={new Date(currentUser.created_at).toLocaleString()}
        />
      </Card>
    );
  };

  const LoginCard = () => {
    const [password, setPassword] = React.useState("");
    return (
      <Card className="mr-4 mt-4 w-[500px] p-4">
        <div className="flex items-center justify-start">
          <FontAwesomeIcon icon={faLock} className="h-5 w-5" />
          <h3 className="ml-4">Authentication</h3>
        </div>
        <Separator className="my-2" />
        <div className="flex items-center justify-between">
          <div className="flex flex-col">
            <p>
              <span className="font-semibold">Email / Password</span>
            </p>
            <p className="text-gray-400">
              {loginAccess.password != null
                ? "Log into Sentinel using your email and password."
                : "Create a password to log into Sentinel."}
            </p>
          </div>
          {loginLoading ? (
            <Button className="ml-auto" variant={"outline"}>
              <Loader2 className="h-6 w-6 animate-spin" />
            </Button>
          ) : (
            <div>
              {loginAccess.password != null ? (
                <Button className="ml-auto" variant={"secondary"}>
                  <span>
                    <FontAwesomeIcon icon={faCheckCircle} className="me-2" />
                  </span>
                  Enabled
                </Button>
              ) : (
                <></>
              )}
            </div>
          )}
        </div>
        {!loginLoading && loginAccess.password == null ? (
          <div className="my-2 flex items-center">
            <Input
              id="gh-username"
              className="mr-2"
              placeholder="Password"
              autoCapitalize="none"
              autoCorrect="off"
              type="password"
              disabled={githubLoading}
              onChange={(e) => {
                setPassword(e.target.value);
              }}
            />
            <Button
              onClick={() => {
                registerPassword(password);
              }}
            >
              Set Password
            </Button>
          </div>
        ) : (
          <></>
        )}
        <div className="mt-2 flex items-center justify-between">
          <div className="flex flex-col">
            <p>
              <span className="font-semibold">OAuth:</span> Discord
            </p>
            <p className="text-gray-400">
              Log into Sentinel using your Discord account.
            </p>
          </div>
          {loginLoading ? (
            <Button className="ml-auto" variant={"outline"}>
              <Loader2 className="h-6 w-6 animate-spin" />
            </Button>
          ) : (
            <div>
              <Button className="ml-auto" variant={"secondary"}>
                <span>
                  <FontAwesomeIcon icon={faCheckCircle} className="me-2" />
                </span>
                Enabled
              </Button>
            </div>
          )}
        </div>
      </Card>
    );
  };

  const DriveCard = () => {
    return (
      <Card className="mr-4 mt-4 w-[500px] p-4">
        <div className="flex items-center justify-start">
          <FontAwesomeIcon icon={faGoogleDrive} className="h-5 w-5" />
          <h3 className="ml-4">Team Drive</h3>
        </div>
        <Separator className="my-2" />
        <div className="flex items-center justify-start">
          <div className="flex flex-col">
            <p>
              <span className="font-semibold">Team Drive:</span> Gaucho Racing
            </p>
            <p className="text-gray-400">
              Access all Gaucho Racing documents through the team's{" "}
              <span
                className="cursor-pointer text-gr-pink hover:text-gr-pink/80"
                onClick={() => window.open(SHARED_DRIVE_URL, "_blank")}
              >
                shared drive
              </span>
              .
            </p>
          </div>
          {driveLoading ? (
            <Button className="ml-auto" variant={"outline"}>
              <Loader2 className="h-6 w-6 animate-spin" />
            </Button>
          ) : (
            <div>
              {driveAccess.role != null ? (
                <Button
                  className="ml-auto"
                  variant={"secondary"}
                  onClick={async () => {
                    await removeUserFromDrive();
                    await addUserToDrive();
                  }}
                >
                  <span>
                    <FontAwesomeIcon icon={faCheckCircle} className="me-2" />
                  </span>
                  Access Granted
                </Button>
              ) : (
                <Button onClick={addUserToDrive}>Request Access</Button>
              )}
            </div>
          )}
        </div>
      </Card>
    );
  };

  const GithubCard = () => {
    const [githubUsername, setGithubUsername] = React.useState("");
    return (
      <Card className="mr-4 mt-4 w-[500px] p-4">
        <div className="flex items-center justify-start">
          <FontAwesomeIcon icon={faGithub} className="h-5 w-5" />
          <h3 className="ml-4">GitHub</h3>
        </div>
        <Separator className="my-2" />
        <div className="flex items-center justify-start">
          <div className="flex flex-col">
            <p>
              <span className="font-semibold">GitHub Org:</span> Gaucho Racing
            </p>
            <p className="text-gray-400">
              Access all Gaucho Racing software through the team's{" "}
              <span
                className="cursor-pointer text-gr-pink hover:text-gr-pink/80"
                onClick={() => window.open(GITHUB_ORG_URL, "_blank")}
              >
                GitHub organization
              </span>
              .
            </p>
          </div>
          {githubLoading ? (
            <Button className="ml-auto" variant={"outline"}>
              <Loader2 className="h-6 w-6 animate-spin" />
            </Button>
          ) : (
            <div>
              {githubAccess.role != null ? (
                <Button className="ml-auto" variant={"secondary"}>
                  <span>
                    <FontAwesomeIcon icon={faCheckCircle} className="me-2" />
                  </span>
                  Access Granted
                </Button>
              ) : (
                <></>
              )}
            </div>
          )}
        </div>
        {!githubLoading && githubAccess.role == null ? (
          <div className="mt-2 flex items-center">
            <Input
              id="gh-username"
              className="mr-2"
              placeholder="GitHub Username"
              autoCapitalize="none"
              autoCorrect="off"
              disabled={githubLoading}
              onChange={(e) => {
                setGithubUsername(e.target.value);
              }}
            />
            <Button
              onClick={() => {
                addUserToGithub(githubUsername);
              }}
            >
              Request Access
            </Button>
          </div>
        ) : (
          <></>
        )}
      </Card>
    );
  };

  return (
    <>
      {authCheckLoading ? (
        <AuthLoading />
      ) : (
        <div className="flex h-screen flex-col justify-between">
          <div className="p-4 lg:p-32 lg:pt-16">
            <h1>Hello {currentUser.first_name}</h1>
            <p className="mt-4 text-gray-400">
              Welcome to Sentinel, Gaucho Racing's central authentication
              service and member directory. Sentinel also provides Single Sign
              On (SSO) access to all our internal services. If you would like to
              build an application using Sentinel, check out our API
              documentation{" "}
              <span
                className="cursor-pointer text-gr-pink hover:text-gr-pink/80"
                onClick={() =>
                  window.open("https://wiki.gauchoracing.com", "_blank")
                }
              >
                here
              </span>
              .
            </p>
            <div className="flex flex-wrap">
              <ProfileCard />
              <div>
                <LoginCard />
                <DriveCard />
                <GithubCard />
              </div>
            </div>
          </div>
          <Footer />
        </div>
      )}
    </>
  );
}

export default App;
