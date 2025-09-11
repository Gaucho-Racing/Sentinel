import { Card } from "@/components/ui/card";
import { GITHUB_ORG_URL, SHARED_DRIVE_URL, WIKI_URL } from "@/consts/config";
import { faGithub } from "@fortawesome/free-brands-svg-icons";
import { faBook, faChartPie, faUsers } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

export default function AppGrid() {
  return (
    <div className="flex flex-col gap-4">
      <div className={`flex flex-row justify-between gap-4`}>
        <Card
          className="flex-1 cursor-pointer p-4 transition-all hover:bg-neutral-800"
          onClick={() => {
            window.open(WIKI_URL, "_blank");
          }}
        >
          <div className="flex flex-col items-center justify-center">
            <FontAwesomeIcon icon={faBook} className="h-10 w-10" />
            <p className="mt-2 text-center text-lg font-semibold">Wiki</p>
          </div>
        </Card>
        <Card
          className="flex-1 cursor-pointer p-4 transition-all hover:bg-neutral-800"
          onClick={() => {
            window.open(SHARED_DRIVE_URL, "_blank");
          }}
        >
          <div className="flex flex-col items-center justify-center">
            <img src={"/logo/apps/drive.png"} className="h-10" />
            <p className="mt-2 text-center text-lg font-semibold">Drive</p>
          </div>
        </Card>
        <Card
          className="flex-1 cursor-pointer p-4 transition-all hover:bg-neutral-800"
          onClick={() => {
            window.open(GITHUB_ORG_URL, "_blank");
          }}
        >
          <div className="flex flex-col items-center justify-center">
            <FontAwesomeIcon icon={faGithub} className="h-10 w-10" />
            <p className="mt-2 text-center text-lg font-semibold">GitHub</p>
          </div>
        </Card>
      </div>
      <div className={`flex flex-row justify-between gap-4`}>
        <Card
          className="flex-1 cursor-pointer p-4 transition-all hover:bg-neutral-800"
          onClick={() => {
            window.location.href = "/users";
          }}
        >
          <div className="flex flex-col items-center justify-center">
            <FontAwesomeIcon icon={faUsers} className="h-10 w-10" />
            <p className="mt-2 text-center text-lg font-semibold">Users</p>
          </div>
        </Card>
        <Card
          className="flex-1 cursor-pointer p-4 transition-all hover:bg-neutral-800"
          onClick={() => {
            window.open(
              "https://portal.singlestore.com?ssoHint=614fcbae-8669-4adb-8a10-3d902ecc4f38",
              "_blank",
            );
          }}
        >
          <div className="flex flex-col items-center justify-center">
            <img src={"/logo/apps/singlestore.png"} className="h-12 w-12" />
            <p className="text-center text-lg font-semibold">SingleStore</p>
          </div>
        </Card>
        <Card
          className="flex-1 cursor-pointer p-4 transition-all hover:bg-neutral-800"
          onClick={() => {
            window.open("https://s2.gauchoracing.com", "_blank");
          }}
        >
          <div className="flex flex-col items-center justify-center">
            <img src={"/logo/apps/s2.png"} className="h-10 w-10" />
            <p className="mt-2 text-center text-lg font-semibold">S2DB</p>
          </div>
        </Card>
        <Card
          className="flex-1 cursor-pointer p-4 transition-all hover:bg-neutral-800"
          onClick={() => {
            window.open("https://portainer.gauchoracing.com", "_blank");
          }}
        >
          <div className="flex flex-col items-center justify-center">
            <img src={"/logo/apps/portainer.png"} className="h-12 w-12" />
            <p className="text-center text-lg font-semibold">Portainer</p>
          </div>
        </Card>
      </div>
      <div className={`flex flex-row justify-between gap-4`}>
        <Card
          className="flex-1 cursor-pointer p-4 transition-all hover:bg-neutral-800"
          onClick={() => {
            window.location.href = "/analytics";
          }}
        >
          <div className="flex flex-col items-center justify-center">
            <FontAwesomeIcon icon={faChartPie} className="h-10 w-10" />
            <p className="mt-2 text-center text-lg font-semibold">Analytics</p>
          </div>
        </Card>
      </div>
    </div>
  );
}
