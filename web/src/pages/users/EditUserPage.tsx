import React from "react";
import axios from "axios";
import { SENTINEL_API_URL, currentUser } from "@/consts/config";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { useNavigate, useParams } from "react-router-dom";
import { Separator } from "@/components/ui/separator";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faArrowLeft, faLock, faUser } from "@fortawesome/free-solid-svg-icons";
import { checkCredentials } from "@/lib/auth";
import Footer from "@/components/Footer";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { User, initUser, setUser } from "@/models/user";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { OutlineButton } from "@/components/ui/outline-button";
import { AuthLoading } from "@/components/AuthLoading";

function EditUserPage() {
  const navigate = useNavigate();
  const { id } = useParams();

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);

  const [canEdit, setCanEdit] = React.useState(false);
  const [editUser, setEditUser] = React.useState<User>(initUser);
  const [userLoading, setUserLoading] = React.useState(false);

  React.useEffect(() => {
    checkAuth().then(() => {
      checkEditPermissions();
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

  const checkEditPermissions = async () => {
    if (currentUser.roles.includes("d_admin") || currentUser.id == id) {
      getUser();
      setCanEdit(true);
    } else {
      setCanEdit(false);
    }
  };

  const getUser = async () => {
    setUserLoading(true);
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/users/${id}`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("token")}`,
        },
      });
      if (response.status == 200) {
        setEditUser(response.data);
      }
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
      setEditUser(initUser);
    }
    setUserLoading(false);
  };

  const saveUser = async () => {
    setUserLoading(true);
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/users/${id}`,
        editUser,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        },
      );
      if (response.status == 200) {
        setEditUser(response.data);
      }
    } catch (error: any) {
      toast(getAxiosErrorMessage(error));
    }
    setUserLoading(false);
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

  const InsufficientPermissionsCard = () => {
    return (
      <Card className="mr-4 mt-4 w-[500px] p-4">
        <div className="flex items-center justify-start">
          <FontAwesomeIcon icon={faLock} className="h-5 w-5" />
          <h3 className="ml-4">Insufficient Permissions</h3>
        </div>
        <Separator className="my-2" />
        <p className="text-gray-400">
          You do not have permission to edit this user's profile. If you believe
          this is an error, please contact an administrator.
        </p>
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
            <div className="mb-4">
              <Button
                variant={"ghost"}
                onClick={() => navigate("/")}
                className="flex items-center"
              >
                <FontAwesomeIcon
                  icon={faArrowLeft}
                  className="mr-2 h-4 w-4 text-gray-400"
                />
                Back to home
              </Button>
            </div>
            <h1>Editing {editUser.first_name}</h1>
            <div className="flex flex-wrap">
              {canEdit ? (
                <Card className="mr-4 mt-4 w-[500px] p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      <FontAwesomeIcon icon={faUser} className="h-5 w-5" />
                      <h3 className="ml-4">Profile</h3>
                    </div>
                    <OutlineButton
                      disabled={userLoading}
                      onClick={async () => {
                        await saveUser();
                        setUser(currentUser, editUser);
                        navigate(`/`);
                      }}
                    >
                      {userLoading && <Loader2 className="mr-2 animate-spin" />}
                      Save Changes
                    </OutlineButton>
                  </div>
                  <Separator className="my-2" />
                  <div className="flex items-center justify-start">
                    <Avatar className="mr-4">
                      <AvatarImage src={editUser.avatar_url} />
                      <AvatarFallback>CN</AvatarFallback>
                    </Avatar>
                    <div className="flex flex-col">
                      <p>
                        {editUser.first_name} {editUser.last_name}
                      </p>
                      <p className="text-gray-400">{editUser.email}</p>
                    </div>
                  </div>
                  <ProfileField label="ID" value={editUser.id} />
                  <ProfileField label="Email" value={editUser.email} />
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">First Name:</div>
                    <Input
                      id="first_name"
                      className=""
                      disabled={userLoading}
                      value={editUser.first_name}
                      onChange={(e) => {
                        setEditUser({
                          ...editUser,
                          first_name: e.target.value,
                        });
                      }}
                    />
                  </div>
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">Last Name:</div>
                    <Input
                      id="last_name"
                      className=""
                      disabled={userLoading}
                      value={editUser.last_name}
                      onChange={(e) => {
                        setEditUser({
                          ...editUser,
                          last_name: e.target.value,
                        });
                      }}
                    />
                  </div>
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">
                      Phone Number:
                    </div>
                    <Input
                      id="phone_number"
                      className=""
                      disabled={userLoading}
                      value={editUser.phone_number}
                      onChange={(e) => {
                        setEditUser({
                          ...editUser,
                          phone_number: e.target.value,
                        });
                      }}
                    />
                  </div>
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">
                      Graduate Level:
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button className="w-full" variant="outline">
                          {editUser.graduate_level}
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="w-[250px]">
                        <DropdownMenuRadioGroup
                          value={editUser.graduate_level}
                          onValueChange={(value) => {
                            setEditUser({
                              ...editUser,
                              graduate_level: value,
                            });
                          }}
                        >
                          <DropdownMenuRadioItem value="Undergraduate">
                            Undergraduate
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="Graduate">
                            Graduate
                          </DropdownMenuRadioItem>
                        </DropdownMenuRadioGroup>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">
                      Graduation Year:
                    </div>
                    <Input
                      id="graduation_year"
                      className=""
                      disabled={userLoading}
                      value={editUser.graduation_year.toString()}
                      onChange={(e) => {
                        const parsedGraduationYear =
                          parseInt(e.target.value) || 0;
                        setEditUser({
                          ...editUser,
                          graduation_year: parsedGraduationYear,
                        });
                      }}
                    />
                  </div>
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">Major:</div>
                    <Input
                      id="major"
                      className=""
                      disabled={userLoading}
                      value={editUser.major}
                      onChange={(e) => {
                        setEditUser({
                          ...editUser,
                          major: e.target.value,
                        });
                      }}
                    />
                  </div>
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">Shirt Size:</div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button className="w-full" variant="outline">
                          {editUser.shirt_size}
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="w-[250px]">
                        <DropdownMenuRadioGroup
                          value={editUser.shirt_size}
                          onValueChange={(value) => {
                            setEditUser({
                              ...editUser,
                              shirt_size: value,
                            });
                          }}
                        >
                          <DropdownMenuRadioItem value="XS">
                            XS
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="S">
                            S
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="M">
                            M
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="L">
                            L
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="XL">
                            XL
                          </DropdownMenuRadioItem>
                        </DropdownMenuRadioGroup>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">Jacket Size:</div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button className="w-full" variant="outline">
                          {editUser.jacket_size}
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="w-[250px]">
                        <DropdownMenuRadioGroup
                          value={editUser.jacket_size}
                          onValueChange={(value) => {
                            setEditUser({
                              ...editUser,
                              jacket_size: value,
                            });
                          }}
                        >
                          <DropdownMenuRadioItem value="XS">
                            XS
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="S">
                            S
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="M">
                            M
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="L">
                            L
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="XL">
                            XL
                          </DropdownMenuRadioItem>
                        </DropdownMenuRadioGroup>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">
                      SAE Member ID:
                    </div>
                    <Input
                      id="sae_registration_number"
                      className=""
                      disabled={userLoading}
                      value={editUser.sae_registration_number}
                      onChange={(e) => {
                        setEditUser({
                          ...editUser,
                          sae_registration_number: e.target.value,
                        });
                      }}
                    />
                  </div>

                  <ProfileField
                    label="Subteams"
                    value={currentUser.subteams
                      .map((subteam) => subteam.name)
                      .join(", ")}
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
              ) : (
                <InsufficientPermissionsCard />
              )}
            </div>
          </div>
          <Footer />
        </div>
      )}
    </>
  );
}

export default EditUserPage;
