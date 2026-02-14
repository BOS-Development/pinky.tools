import Document, { Head, Html, Main, NextScript } from "next/document";

export default class MyDocument extends Document {
  render() {
    return (
      <Html>
        <Head />
        <body style={{ backgroundColor: "#161b22", color: "white" }}>
          <Main />
          <NextScript />
        </body>
      </Html>
    );
  }
}
