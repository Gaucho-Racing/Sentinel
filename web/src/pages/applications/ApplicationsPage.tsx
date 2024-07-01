import React from "react";
import axios from "axios";
import { SENTINEL_API_URL, currentUser } from "@/consts/config";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Loader2, User } from "lucide-react";
import { toast } from "sonner";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { useNavigate, useParams } from "react-router-dom";
import { checkCredentials } from "@/lib/auth";
import Footer from "@/components/Footer";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { ClientApplication, initClientApplication } from "@/models/application";
import { OutlineButton } from "@/components/ui/outline-button";
import { AuthLoading } from "@/components/AuthLoading";
import { User, initUser } from "@/models/user";
import { Input } from "@/components/ui/input";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTrash } from "@fortawesome/free-solid-svg-icons";

function ApplicationsPage() {
  const navigate = useNavigate();
  const { id } = useParams();

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);

  const [applicationsLoading, setApplicationsLoading] = React.useState(false);
  const [applications, setApplications] = React.useState<ClientApplication[]>(
    [],
  );
  const [selectedApplication, setSelectedApplication] =
    React.useState<ClientApplication>(initClientApplication);
  const [selectedApplicationLoading, setSelectedApplicationLoading] =
    React.useState(false);
  const [selectedOwner, setSelectedOwner] = React.useState<User>(initUser);
  const [canEdit, setCanEdit] = React.useState(false);

  const [creatingApplication, setCreatingApplication] = React.useState(false);

  React.useEffect(() => {
    checkAuth().then(async () => {
      await getApplications();
      init();
    });
  }, [id]);

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

  const init = async () => {
    if (id) {
      if (id == "new") {
        setSelectedApplication(initClientApplication);
        setSelectedOwner(initUser);
        setCreatingApplication(true);
      } else {
        await getSelectedApplication(id);
      }
    }
  };

  const getSelectedApplication = async (applicationId: string) => {
    setSelectedApplicationLoading(true);
    try {
      const response = await axios.get(
        `${SENTINEL_API_URL}/applications/${applicationId}`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      setSelectedApplication(response.data);
      getUser(response.data.user_id);
      if (
        response.data.user_id == currentUser.id ||
        currentUser.roles.includes("d_admin")
      ) {
        setCanEdit(true);
      }
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
      navigate("/applications");
    }
    setSelectedApplicationLoading(false);
  };

  const getApplications = async () => {
    if (applications.length == 0) {
      setApplicationsLoading(true);
    }
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/applications`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("token")}`,
        },
      });
      setApplications(response.data);
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
    setApplicationsLoading(false);
  };

  const getUser = async (userId: string) => {
    setSelectedOwner(initUser);
    setSelectedApplicationLoading(true);
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/users/${userId}`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("token")}`,
        },
      });
      setSelectedOwner(response.data);
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
    setSelectedApplicationLoading(false);
  };

  const createApplication = async () => {
    if (selectedApplication.name.trim() == "") {
      toast("You must specify a name for your application.");
      return;
    }
    const cleanedApplication = {
      ...selectedApplication,
      redirect_uris: selectedApplication.redirect_uris.filter(
        (uri) => uri.trim() !== "",
      ),
    };
    setSelectedApplication(cleanedApplication);
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/applications`,
        cleanedApplication,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      if (response.status == 201) {
        navigate(`/applications/${response.data.id}`);
      }
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
    getApplications();
  };

  const ApplicationListItem = (props: { application: ClientApplication }) => {
    return (
      <Card className="mt-2 p-2">
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

  const ApplicationLoadingCard = () => {
    return (
      <Card className="w-full p-4 md:p-16">
        <div className="flex w-full justify-center">
          <Loader2 className="animate-spin" />
        </div>
      </Card>
    );
  };

  return (
    <>
      {authCheckLoading ? (
        <AuthLoading />
      ) : (
        <div className="flex h-screen flex-col justify-between">
          <div className="flex-grow p-4 lg:p-32 lg:pt-16">
            <h1>Applications</h1>
            <p className="mt-4 text-gray-400">
              Want to create a new team application? Use the Sentinel API to
              easily authenticate Gaucho Racing members. Check out our API
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
            <div className="mt-2 flex h-full flex-col lg:flex-row">
              <div className="mb-4 w-full overflow-y-auto lg:mb-0 lg:mr-4 lg:w-1/3">
                {applicationsLoading ? (
                  <div className="flex w-full justify-center p-4">
                    <Loader2 className="animate-spin" />
                  </div>
                ) : (
                  <div>
                    {applications.map((application) => (
                      <div key={application.id} className="mt-2">
                        <ApplicationListItem application={application} />
                      </div>
                    ))}
                    <OutlineButton
                      onClick={() => navigate("/applications/new")}
                      className="mt-2"
                    >
                      New Application
                    </OutlineButton>
                  </div>
                )}
              </div>
              <div className="mt-2 w-full overflow-y-auto lg:w-2/3">
                {selectedApplicationLoading ? (
                  <ApplicationLoadingCard />
                ) : (
                  <Card className="w-full p-4 md:p-8">
                    <div className="flex flex-col items-start">
                      {canEdit ? (
                        <Input
                          id="name"
                          className="w-full text-2xl font-semibold"
                          placeholder="Application Name"
                          value={selectedApplication.name}
                          onChange={(e) => {
                            setSelectedApplication({
                              ...selectedApplication,
                              name: e.target.value,
                            });
                          }}
                        />
                      ) : (
                        <h1 className="text-2xl font-semibold">
                          {selectedApplication.name}
                        </h1>
                      )}
                      <p className="mt-2 text-gray-400"></p>
                      <div className="mx-2 mt-2 flex w-full items-center">
                        <div className="mr-2 w-1/5 font-semibold">
                          Client ID:
                        </div>
                        <Input
                          id="client_id"
                          className="w-4/5"
                          disabled={true}
                          value={selectedApplication.id}
                        />
                      </div>
                      <div className="mx-2 mt-2 flex w-full items-center">
                        <div className="mr-2 w-1/5 font-semibold">
                          Client Secret:
                        </div>
                        <Input
                          id="client_secret"
                          className="w-4/5"
                          disabled={canEdit}
                          value={selectedApplication.secret}
                          type={canEdit ? "text" : "password"}
                        />
                      </div>
                      <div className="mx-2 mt-2 flex w-full items-center">
                        <div className="mr-2 w-1/5 font-semibold">Owner:</div>
                        <div className="mt-2 flex items-center">
                          <Avatar className="mr-4">
                            <AvatarImage src={selectedOwner.avatar_url} />
                            <AvatarFallback>CN</AvatarFallback>
                          </Avatar>
                          <div className="flex flex-col items-start justify-center">
                            <div>
                              {selectedOwner.first_name}{" "}
                              {selectedOwner.last_name}
                            </div>
                            <div className="text-gray-400">
                              {selectedOwner.email}
                            </div>
                          </div>
                        </div>
                      </div>
                      <div className="mx-2 mt-2 flex w-full flex-col items-start">
                        <div className="font-semibold">Redirect URIs:</div>
                        <p className="mt-1 text-gray-400">
                          You must specify at least one URI for authentication
                          to work. If you pass a URI in an OAuth request, it
                          must exactly match one of the URIs you enter here.
                        </p>
                        {selectedApplication.redirect_uris.map((uri, index) => (
                          <div
                            key={index}
                            className="mt-2 flex w-full items-center"
                          >
                            <div className="mr-2 w-1/5 font-semibold">
                              URI {index + 1}:
                            </div>
                            <Input
                              id={`redirect_uri_${index}`}
                              className="w-4/5"
                              disabled={!canEdit}
                              value={uri}
                              onChange={(e) => {
                                const newUris =
                                  selectedApplication.redirect_uris;
                                newUris[index] = e.target.value;
                                setSelectedApplication({
                                  ...selectedApplication,
                                  redirect_uris: newUris,
                                });
                              }}
                            />
                            <Button
                              variant="ghost"
                              size="icon"
                              className="ml-2"
                              onClick={() => {
                                const newUris =
                                  selectedApplication.redirect_uris;
                                newUris.splice(index, 1);
                                setSelectedApplication({
                                  ...selectedApplication,
                                  redirect_uris: newUris,
                                });
                              }}
                            >
                              <FontAwesomeIcon
                                icon={faTrash}
                                className="h-4 w-4 text-destructive"
                              />
                            </Button>
                          </div>
                        ))}
                        {canEdit && (
                          <div className="mt-4 flex w-full items-center justify-between">
                            <Button
                              variant={"outline"}
                              onClick={() => {
                                const newUris =
                                  selectedApplication.redirect_uris;
                                newUris.push("");
                                setSelectedApplication({
                                  ...selectedApplication,
                                  redirect_uris: newUris,
                                });
                              }}
                              className="py-5"
                            >
                              Add Redirect
                            </Button>
                            <div className="flex items-center justify-end">
                              <Button
                                variant={"outline"}
                                onClick={() => {
                                  getSelectedApplication(
                                    selectedApplication.id,
                                  );
                                }}
                                className="mr-2 py-5"
                              >
                                Discard Changes
                              </Button>
                              <OutlineButton
                                onClick={() => {
                                  createApplication();
                                }}
                              >
                                Save Changes
                              </OutlineButton>
                            </div>
                          </div>
                        )}
                      </div>
                    </div>
                  </Card>
                )}
              </div>
            </div>
          </div>
          <Footer />
        </div>
      )}
    </>
  );
}

export default ApplicationsPage;
