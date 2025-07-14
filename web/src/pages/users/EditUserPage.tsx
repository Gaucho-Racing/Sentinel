import React from "react";
import axios from "axios";
import { SENTINEL_API_URL } from "@/consts/config";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { useNavigate, useParams } from "react-router-dom";
import { Separator } from "@/components/ui/separator";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faArrowLeft, faLock, faUser } from "@fortawesome/free-solid-svg-icons";
import { checkCredentials } from "@/lib/auth";
import Footer from "@/components/Footer";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { User, initUser, isInnerCircle } from "@/models/user";
import { setUser, useUser, getUser as getCurrentUser } from "@/lib/store";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { OutlineButton } from "@/components/ui/outline-button";
import { AuthLoading } from "@/components/AuthLoading";
import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { Calendar } from "@/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { CalendarIcon, Loader2 } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { notify } from "@/lib/notify";

function EditUserPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const currentUser = useUser();

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);

  const [canEdit, setCanEdit] = React.useState(false);
  const [editUser, setEditUser] = React.useState<User>(initUser);
  const [userLoading, setUserLoading] = React.useState(false);

  const [date, setDate] = React.useState<Date>();

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
    const currentUser = getCurrentUser();
    if (isInnerCircle(currentUser) || currentUser.id == id) {
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
          Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
        },
      });
      if (response.status == 200) {
        setEditUser(response.data);
        if (response.data.birthday) {
          const date = new Date(response.data.birthday);
          setDate(date);
        }
      }
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
      setEditUser(initUser);
    }
    setUserLoading(false);
  };

  const saveUser = async () => {
    setUserLoading(true);
    if (date) {
      const dateString = date.toLocaleDateString("en-US", {
        year: "numeric",
        month: "long",
        day: "numeric",
      });
      editUser.birthday = dateString;
    }
    try {
      const response = await axios.post(
        `${SENTINEL_API_URL}/users/${id}`,
        editUser,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
          },
        },
      );
      if (response.status == 200) {
        setEditUser(response.data);
      }
      notify.success(
        "Changes saved",
        "Your profile has successfully been updated.",
      );
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
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
                        if (currentUser.id == id) {
                          setUser(editUser);
                        }
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
                    <div className="mr-2 w-2/3 font-semibold">Gender:</div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button className="w-full" variant="outline">
                          {editUser.gender}
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="w-[250px]">
                        <DropdownMenuRadioGroup
                          value={editUser.gender}
                          onValueChange={(value) => {
                            setEditUser({
                              ...editUser,
                              gender: value,
                            });
                          }}
                        >
                          <DropdownMenuRadioItem value="Male">
                            Male
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="Female">
                            Female
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value="Other">
                            Other
                          </DropdownMenuRadioItem>
                        </DropdownMenuRadioGroup>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                  <div className="mx-2 mt-2 flex items-center">
                    <div className="mr-2 w-2/3 font-semibold">Birthday:</div>
                    <Popover>
                      <PopoverTrigger asChild>
                        <Button
                          variant={"outline"}
                          className={cn(
                            "w-full justify-start text-left font-normal",
                            !date && "text-muted-foreground",
                          )}
                        >
                          <CalendarIcon className="mr-2 h-4 w-4" />
                          {date ? (
                            format(date, "PPP")
                          ) : (
                            <span>Pick a date</span>
                          )}
                        </Button>
                      </PopoverTrigger>
                      <PopoverContent className="w-auto bg-background p-0">
                        <div className="flex justify-between">
                          <div className="w-full p-2">
                            <Select
                              value={date ? date.getFullYear().toString() : ""}
                              onValueChange={(value) => {
                                const newDate = new Date(date || new Date());
                                newDate.setFullYear(parseInt(value));
                                setDate(newDate);
                              }}
                            >
                              <SelectTrigger className="w-full">
                                <SelectValue placeholder="Select year" />
                              </SelectTrigger>
                              <SelectContent>
                                {Array.from({ length: 100 }, (_, i) => {
                                  const year = new Date().getFullYear() - i;
                                  return (
                                    <SelectItem
                                      key={year}
                                      value={year.toString()}
                                    >
                                      {year}
                                    </SelectItem>
                                  );
                                })}
                              </SelectContent>
                            </Select>
                          </div>
                          <div className="w-full p-2">
                            <Select
                              value={
                                date ? (date.getMonth() + 1).toString() : ""
                              }
                              onValueChange={(value) => {
                                const newDate = new Date(date || new Date());
                                newDate.setMonth(parseInt(value) - 1);
                                setDate(newDate);
                              }}
                            >
                              <SelectTrigger className="w-full">
                                <SelectValue placeholder="Select month" />
                              </SelectTrigger>
                              <SelectContent>
                                {Array.from({ length: 12 }, (_, i) => {
                                  const month = i + 1;
                                  return (
                                    <SelectItem
                                      key={month}
                                      value={month.toString()}
                                    >
                                      {new Date(0, i).toLocaleString(
                                        "default",
                                        { month: "long" },
                                      )}
                                    </SelectItem>
                                  );
                                })}
                              </SelectContent>
                            </Select>
                          </div>
                        </div>
                        <Calendar
                          mode="single"
                          selected={date}
                          onSelect={setDate}
                          month={date || new Date()}
                          initialFocus
                        />
                      </PopoverContent>
                    </Popover>
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
                    value={editUser.subteams
                      .map((subteam) => subteam.name)
                      .join(", ")}
                  />
                  <div className="mx-2 mt-2 flex">
                    <div className="mr-2 font-semibold">Roles:</div>
                    <div className="flex flex-wrap">
                      {editUser.roles.map((role) => (
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
                    value={new Date(editUser.updated_at).toLocaleString()}
                  />
                  <ProfileField
                    label="Created At"
                    value={new Date(editUser.created_at).toLocaleString()}
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
