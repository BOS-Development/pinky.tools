import { SessionProvider } from "next-auth/react";
import ThemeRegistry from "@industry-tool/components/ThemeRegistry";

export default function App({
  Component,
  pageProps: { session, ...pageProps },
}) {
  return (
    <ThemeRegistry>
      <SessionProvider session={session}>
        <Component {...pageProps} />
      </SessionProvider>
    </ThemeRegistry>
  );
}
