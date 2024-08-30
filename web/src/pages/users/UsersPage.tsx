import React from "react";
import axios from "axios";
import { SENTINEL_API_URL } from "@/consts/config";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getAxiosErrorMessage } from "@/lib/axios-error-handler";
import { useNavigate, useParams } from "react-router-dom";
import { Separator } from "@/components/ui/separator";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faArrowLeft, faLock, faUser } from "@fortawesome/free-solid-svg-icons";
import { checkCredentials } from "@/lib/auth";
import Footer from "@/components/Footer";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { User, initUser } from "@/models/user";
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
import { Badge } from "@/components/ui/badge";

function UsersPage() {
  const navigate = useNavigate();
  const currentUser = useUser();

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);

  const [displayUsers, setDisplayUsers] = React.useState<User[]>([]);
  const [users, setUsers] = React.useState<User[]>([]);
  const [userLoading, setUserLoading] = React.useState(false);

  const [date, setDate] = React.useState<Date>();

  React.useEffect(() => {
    checkAuth().then(() => {
      getUsers();
    });
  }, [currentUser]);

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

  const getUsers = async () => {
    setUserLoading(true);
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/users`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
        },
      });
      if (response.status == 200) {
        const sortedUsers = response.data.sort((a: User, b: User) =>
          a.first_name.localeCompare(b.first_name),
        );
        setUsers(sortedUsers);
        setDisplayUsers(sortedUsers);
      }
    } catch (error: any) {
      notify.error(getAxiosErrorMessage(error));
    }
    setUserLoading(false);
  };

  const UserCard = ({ user }: { user: User }) => {
    return (
      <Card className="w-64">
        <CardHeader className="flex flex-row items-center space-x-4 pb-2">
          <Avatar>
            <AvatarImage src={user.avatar_url} alt={user.first_name} />
            <AvatarFallback>
              {user.first_name[0]}
              {user.last_name[0]}
            </AvatarFallback>
          </Avatar>
          <CardTitle>
            {user.first_name} {user.last_name}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="mb-2 text-sm text-muted-foreground">{user.email}</p>
          <div className="flex flex-wrap">
            {user.roles.map((role) => (
              <div key={role} className="mx-1 mb-2">
                <Card className="rounded-sm px-1 text-gray-400">
                  <code className="">{role}</code>
                </Card>
              </div>
            ))}
          </div>
        </CardContent>
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
            <h1>Users</h1>
            <div className="flex flex-wrap gap-2">
              {displayUsers.map((user) => (
                <UserCard key={user.id} user={user} />
              ))}
            </div>
          </div>
          <Footer />
        </div>
      )}
    </>
  );
}

export default UsersPage;
