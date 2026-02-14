import { getServerSession } from "next-auth/next";

export default async function Home() {
  const session = await getServerSession();

  console.log(session);

  return <div>Hello {session?.user?.name}</div>;
}
