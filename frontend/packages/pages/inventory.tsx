import { useSession } from "next-auth/react";

export default function Inventory() {
  const { data: session, status } = useSession();

  if (status === "authenticated") {
    console.log(session);
    return <p>Signed in as {session?.user?.name}</p>;
  }

  return <div>not logged in</div>;
}
