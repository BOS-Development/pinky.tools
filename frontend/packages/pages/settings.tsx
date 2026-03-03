import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import DiscordSettings from "@industry-tool/components/settings/DiscordSettings";

export default function Settings() {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return (
    <>
      <Navbar />
      <div className="max-w-3xl mx-auto px-4 py-8">
        <h1 className="text-2xl font-display font-semibold mb-6">Settings</h1>
        <DiscordSettings />
      </div>
    </>
  );
}
