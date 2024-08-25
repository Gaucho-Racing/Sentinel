import React from "react";
import axios from "axios";
import {
  GITHUB_ORG_URL,
  SENTINEL_API_URL,
  SHARED_DRIVE_URL,
  WIKI_URL,
} from "@/consts/config";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { useNavigate } from "react-router-dom";
import { Separator } from "@/components/ui/separator";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faArrowUpRightFromSquare,
  faBook,
  faCheckCircle,
  faLock,
  faUser,
} from "@fortawesome/free-solid-svg-icons";
import { checkCredentials, logout, saveAccessToken } from "@/lib/auth";
import Footer from "@/components/Footer";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import {
  faAppStore,
  faGithub,
  faGoogleDrive,
} from "@fortawesome/free-brands-svg-icons";
import { ClientApplication } from "@/models/application";
import { OutlineButton } from "@/components/ui/outline-button";
import { AuthLoading } from "@/components/AuthLoading";
import { notify } from "@/lib/notify";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { getUser, useUser } from "@/lib/store";

function App() {
  const navigate = useNavigate();
  const currentUser = useUser();

  const [cardWidth, setCardWidth] = React.useState(500);

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);

  const [loginLoading, setLoginLoading] = React.useState(false);
  const [loginAccess, setLoginAccess] = React.useState<any>({});

  const [driveLoading, setDriveLoading] = React.useState(false);
  const [driveAccess, setDriveAccess] = React.useState<any>({});

  const [githubLoading, setGithubLoading] = React.useState(false);
  const [githubAccess, setGithubAccess] = React.useState<any>({});

  const [applicationsLoading, setApplicationsLoading] = React.useState(false);
  const [applications, setApplications] = React.useState<ClientApplication[]>(
    [],
  );

  const handleResize = () => {
    const width = window.innerWidth;
    if (width < 600) {
      setCardWidth(width - 32);
    } else {
      setCardWidth(500);
    }
  };

  React.useEffect(() => {
    window.addEventListener("resize", handleResize);
    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  React.useEffect(() => {
    checkAuth().then(() => {
      checkLoginAccess();
      checkDriveAccess();
      checkGithubAccess();
      getApplications();
    });
  }, []);

  const checkAuth = async () => {
    setAuthCheckLoading(true);
    const currentRoute = window.location.pathname + window.location.search;
    const status = await checkCredentials();
    if (status != 0) {
      if (currentRoute == "/") {
        navigate(`/auth/login`);
      } else {
        navigate(`/auth/login?route=${encodeURIComponent(currentRoute)}`);
      }
    } else {
      setAuthCheckLoading(false);
    }
  };

  const checkLoginAccess = async () => {
    let currentUser = getUser();
    setLoginLoading(true);
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/users/${currentUser.id}/auth`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      setLoginAccess(response.data);
    } catch (error: any) {
      if (!getAxiosErrorMessage(error).includes("No authentication found")) {
        notify.error(getAxiosErrorMessage(error));
      }
    }
    setLoginLoading(false);
  };

  const registerPassword = async (password: string) => {
    let currentUser = getUser();
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
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      if (response.status == 200) {
        saveAccessToken(response.data.access_token);
        checkCredentials();
      }
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    checkLoginAccess();
  };

  const resetPassword = async () => {
    let currentUser = getUser();
    setLoginLoading(true);
    try {
      const response = await axios.delete(
        `${SENTINEL_API_URL}/users/${currentUser.id}/auth`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      if (response.status == 200) {
        window.location.reload();
      }
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    checkLoginAccess();
  };

  const checkDriveAccess = async () => {
    let currentUser = getUser();
    setDriveLoading(true);
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/users/${currentUser.id}/drive`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      setDriveAccess(response.data);
    } catch (error: any) {
      if (!getAxiosErrorMessage(error).includes("No permissions found")) {
        notify.error(getAxiosErrorMessage(error));
      }
    }
    setDriveLoading(false);
  };

  const addUserToDrive = async () => {
    let currentUser = getUser();
    setDriveLoading(true);
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/users/${currentUser.id}/drive`,
        {},
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      setDriveAccess(response.data);
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    checkDriveAccess();
  };

  const removeUserFromDrive = async () => {
    let currentUser = getUser();
    setDriveLoading(true);
    try {
      const response = await axios.delete(
        `${SENTINEL_API_URL}/users/${currentUser.id}/drive`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      setDriveAccess(response.data);
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    checkDriveAccess();
  };

  const checkGithubAccess = async () => {
    let currentUser = getUser();
    setGithubLoading(true);
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/users/${currentUser.id}/github`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      setGithubAccess(response.data);
    } catch (error: any) {
      if (!getAxiosErrorMessage(error).includes("user does not have")) {
        notify.error(getAxiosErrorMessage(error));
      }
    }
    setGithubLoading(false);
  };

  const addUserToGithub = async (username: string) => {
    let currentUser = getUser();
    setGithubLoading(true);
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/users/${currentUser.id}/github`,
        {
          username: username,
        },
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      setGithubAccess(response.data);
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    checkGithubAccess();
  };

  const getApplications = async () => {
    let currentUser = getUser();
    setApplicationsLoading(true);
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/users/${currentUser.id}/applications`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      setApplications(response.data);
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    setApplicationsLoading(false);
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
      <Card className={`mr-4 mt-4 w-[${cardWidth}px] p-4`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center">
            <FontAwesomeIcon icon={faUser} className="h-5 w-5" />
            <h3 className="ml-4">Profile</h3>
          </div>
          <Button
            variant={"outline"}
            onClick={() => navigate(`/users/${currentUser.id}/edit`)}
          >
            Edit
          </Button>
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
        <ProfileField label="Gender" value={currentUser.gender} />
        <ProfileField label="Birthday" value={currentUser.birthday} />
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
      <Card className={`mr-4 mt-4 w-[${cardWidth}px] p-4`}>
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
            <div>
              {loginAccess.password != null ? (
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <p className="cursor-pointer text-gr-pink">
                      Reset password?
                    </p>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                      <AlertDialogDescription>
                        You will no longer be able to sign into Sentinel using
                        your email and password.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={resetPassword}
                        className="bg-destructive text-destructive-foreground hover:bg-destructive/50"
                      >
                        Reset
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              ) : (
                <></>
              )}
            </div>
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
            <OutlineButton
              onClick={() => {
                registerPassword(password);
              }}
            >
              Set Password
            </OutlineButton>
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

  const WikiCard = () => {
    return (
      <Card className={`mr-4 mt-4 w-[${cardWidth}px] p-4`}>
        <div className="flex items-center justify-start">
          <FontAwesomeIcon icon={faBook} className="h-5 w-5" />
          <h3 className="ml-4">Wiki</h3>
        </div>
        <Separator className="my-2" />
        <div className="flex items-center justify-start">
          <div className="flex flex-col">
            {loginAccess.password != null ? (
              <p>Login with your Sentinel account.</p>
            ) : (
              <p>Please set a password for your sentinel account first!</p>
            )}
            <p className="text-gray-400">
              Access all Gaucho Racing documentation through the team's{" "}
              <span
                className="cursor-pointer text-gr-pink hover:text-gr-pink/80"
                onClick={() => window.open(WIKI_URL, "_blank")}
              >
                wiki
              </span>
              .
            </p>
          </div>
          {loginLoading ? (
            <Button className="ml-auto" variant={"outline"}>
              <Loader2 className="h-6 w-6 animate-spin" />
            </Button>
          ) : (
            <div>
              {loginAccess.password != null ? (
                <OutlineButton
                  className="ml-auto"
                  onClick={() => window.open(WIKI_URL, "_blank")}
                >
                  <span>
                    <FontAwesomeIcon
                      icon={faArrowUpRightFromSquare}
                      className="me-2"
                    />
                  </span>
                  Launch Wiki
                </OutlineButton>
              ) : (
                <></>
              )}
            </div>
          )}
        </div>
      </Card>
    );
  };

  const DriveCard = () => {
    return (
      <Card className={`mr-4 mt-4 w-[${cardWidth}px] p-4`}>
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
                <OutlineButton onClick={addUserToDrive}>
                  Request Access
                </OutlineButton>
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
      <Card className={`mr-4 mt-4 w-[${cardWidth}px] p-4`}>
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
            <OutlineButton
              onClick={() => {
                addUserToGithub(githubUsername);
              }}
            >
              Request Access
            </OutlineButton>
          </div>
        ) : (
          <></>
        )}
      </Card>
    );
  };

  const ApplicationsCard = () => {
    return (
      <Card className={`mr-4 mt-4 w-[${cardWidth}px] p-4`}>
        <div className="flex items-center justify-start">
          <FontAwesomeIcon icon={faAppStore} className="h-5 w-5" />
          <h3 className="ml-4">My Applications</h3>
        </div>
        <Separator className="my-2" />
        <div className="items-center justify-start">
          {applicationsLoading ? (
            <div className="flex w-full justify-center p-4">
              <Loader2 className="animate-spin" />
            </div>
          ) : (
            <div>
              {applications.length > 0 ? (
                applications.map((application) => (
                  <div key={application.id} className="mt-2">
                    <ApplicationListItem application={application} />
                  </div>
                ))
              ) : (
                <NoApplicationsCard />
              )}
            </div>
          )}
        </div>
      </Card>
    );
  };

  const ApplicationListItem = (props: { application: ClientApplication }) => {
    return (
      <Card className="border-none">
        <div className="flex items-center justify-between">
          <div className="items-center">
            <div className="mr-2 font-semibold">{props.application.name}</div>
            <div className="text-gray-400">
              Client ID: {props.application.id}
            </div>
          </div>
          <Button
            variant={"outline"}
            onClick={() => navigate(`/applications/${props.application.id}`)}
          >
            View
          </Button>
        </div>
      </Card>
    );
  };

  const NoApplicationsCard = () => {
    return (
      <div className="flex w-full flex-col items-center justify-center">
        <p className="mt-2 text-gray-400">
          You have not created any applications yet.
        </p>
        <OutlineButton
          onClick={() => navigate("/applications/new")}
          className="mt-4"
        >
          Create Application
        </OutlineButton>
      </div>
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
                  window.open(
                    "https://wiki.gauchoracing.com/books/sentinel/page/api-documentation",
                    "_blank",
                  )
                }
              >
                here
              </span>
              .
            </p>
            <div className="flex flex-wrap">
              <div>
                <ProfileCard />
                <div className={`mr-4 mt-4 w-[${cardWidth}px]`}>
                  <Button
                    className="w-full"
                    variant={"destructive"}
                    onClick={() => {
                      logout();
                      navigate("/auth/login");
                    }}
                  >
                    Sign Out
                  </Button>
                </div>
              </div>
              <div>
                <LoginCard />
                <WikiCard />
                <DriveCard />
                <GithubCard />
              </div>
              <div>
                <ApplicationsCard />
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
