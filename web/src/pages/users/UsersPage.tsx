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
import {
  faArrowLeft,
  faArrowRight,
  faChevronRight,
  faEnvelope,
  faLock,
  faUser,
} from "@fortawesome/free-solid-svg-icons";
import { checkCredentials } from "@/lib/auth";
import Footer from "@/components/Footer";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { User, initUser } from "@/models/user";
import { setUser, useUser, getUser as getCurrentUser } from "@/lib/store";
import { AuthLoading } from "@/components/AuthLoading";
import Fuse from "fuse.js";
import { faDiscord } from "@fortawesome/free-brands-svg-icons";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { notify } from "@/lib/notify";

function UsersPage() {
  const navigate = useNavigate();
  const currentUser = useUser();

  const [authCheckLoading, setAuthCheckLoading] = React.useState(false);

  const [displayUsers, setDisplayUsers] = React.useState<User[]>([]);
  const [users, setUsers] = React.useState<User[]>([]);
  const [userLoading, setUserLoading] = React.useState(false);

  const [date, setDate] = React.useState<Date>();

  const [searchTerm, setSearchTerm] = React.useState("");
  const [selectedSubteam, setSelectedSubteam] = React.useState("all");
  const [selectedRole, setSelectedRole] = React.useState("all");

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

  const filterUsers = () => {
    let filteredUsers = users;

    if (searchTerm) {
      const fuse = new Fuse(filteredUsers, {
        keys: ["first_name", "last_name", "email", "username"],
        threshold: 0.3,
      });
      filteredUsers = fuse.search(searchTerm).map((result) => result.item);
    }

    if (selectedSubteam && selectedSubteam !== "all") {
      filteredUsers = filteredUsers.filter((user) =>
        user.subteams.some((subteam) => subteam.name === selectedSubteam),
      );
    }

    if (selectedRole && selectedRole !== "all") {
      filteredUsers = filteredUsers.filter((user) =>
        user.roles.includes(selectedRole),
      );
    }

    setDisplayUsers(filteredUsers);
  };

  React.useEffect(() => {
    filterUsers();
  }, [searchTerm, selectedSubteam, selectedRole, users]);

  const allSubteams = React.useMemo(() => {
    const subteamSet = new Set<string>();
    users.forEach((user) =>
      user.subteams.forEach((subteam) => subteamSet.add(subteam.name)),
    );
    return Array.from(subteamSet);
  }, [users]);

  const allRoles = React.useMemo(() => {
    const roleSet = new Set<string>();
    users.forEach((user) =>
      user.roles
        .filter((role) => role.startsWith("d_"))
        .forEach((role) => roleSet.add(role)),
    );
    return Array.from(roleSet);
  }, [users]);

  const formatRoleName = (role: string) => {
    return role
      .slice(2)
      .split("_")
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(" ");
  };

  const UserCard = ({ user }: { user: User }) => {
    return (
      <Card className="w-1/3 p-4 transition-all hover:cursor-pointer hover:bg-neutral-800">
        <div className="flex items-center justify-between">
          <div className="flex flex-col items-start justify-start">
            <div className="flex items-center space-x-2">
              <Avatar>
                <AvatarImage src={user.avatar_url} alt={user.first_name} />
                <AvatarFallback>
                  {user.first_name[0]}
                  {user.last_name[0]}
                </AvatarFallback>
              </Avatar>
              <h3>
                {user.first_name} {user.last_name}
              </h3>
            </div>
            <div className="mt-2 flex space-x-4">
              <div className="flex items-center space-x-2">
                <FontAwesomeIcon icon={faEnvelope} className="" size="lg" />
                <p className="text-gray-400">{user.email}</p>
              </div>
              <div className="flex items-center space-x-2">
                <FontAwesomeIcon icon={faDiscord} className="" size="lg" />
                <p className="text-gray-400">{user.username}</p>
              </div>
            </div>
            <div className="mt-2 flex space-x-2">
              <div className="font-semibold">Subteams:</div>
              <div className="flex flex-wrap gap-2">
                {user.subteams.map((subteam) => (
                  <div key={subteam.id}>
                    <Card className="rounded-sm px-1 text-gray-400">
                      <code className="">{subteam.name}</code>
                    </Card>
                  </div>
                ))}
              </div>
            </div>
            <div className="mt-2 flex space-x-2">
              <div className="font-semibold">Roles:</div>
              <div className="flex flex-wrap gap-2">
                {user.roles
                  .filter((role) => role.startsWith("d_"))
                  .map((role) => (
                    <div key={role}>
                      <Card className="rounded-sm px-1 text-gray-400">
                        <code className="">{formatRoleName(role)}</code>
                      </Card>
                    </div>
                  ))}
              </div>
            </div>
          </div>
          <FontAwesomeIcon icon={faChevronRight} className="text-gray-400" />
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
            <div className="mt-4 flex flex-col">
              <SearchAndFilterComponent
                searchTerm={searchTerm}
                setSearchTerm={setSearchTerm}
                selectedSubteam={selectedSubteam}
                setSelectedSubteam={setSelectedSubteam}
                selectedRole={selectedRole}
                setSelectedRole={setSelectedRole}
                allSubteams={allSubteams}
                allRoles={allRoles}
                formatRoleName={formatRoleName}
              />
            </div>
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

interface SearchAndFilterComponentProps {
  searchTerm: string;
  setSearchTerm: (value: string) => void;
  selectedSubteam: string;
  setSelectedSubteam: (value: string) => void;
  selectedRole: string;
  setSelectedRole: (value: string) => void;
  allSubteams: string[];
  allRoles: string[];
  formatRoleName: (role: string) => string;
}

const SearchAndFilterComponent: React.FC<SearchAndFilterComponentProps> = ({
  searchTerm,
  setSearchTerm,
  selectedSubteam,
  setSelectedSubteam,
  selectedRole,
  setSelectedRole,
  allSubteams,
  allRoles,
  formatRoleName,
}) => {
  return (
    <div className="mb-4 flex flex-wrap gap-4">
      <Input
        placeholder="Search users..."
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
        className="w-full md:w-64"
      />
      <Select value={selectedSubteam} onValueChange={setSelectedSubteam}>
        <SelectTrigger className="w-full md:w-48">
          <SelectValue placeholder="Filter by subteam" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Subteams</SelectItem>
          {allSubteams.map((subteam) => (
            <SelectItem key={subteam} value={subteam}>
              {subteam}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select value={selectedRole} onValueChange={setSelectedRole}>
        <SelectTrigger className="w-full md:w-48">
          <SelectValue placeholder="Filter by role" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Roles</SelectItem>
          {allRoles.map((role) => (
            <SelectItem key={role} value={role}>
              {formatRoleName(role)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
};
