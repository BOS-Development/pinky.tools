import Head from "next/head";
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import ContactsList from "@industry-tool/components/contacts/ContactsList";

export default function Contacts() {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return (
    <>
      <Head><title>Contacts — pinky.tools</title></Head>
      <ContactsList />
    </>
  );
}
