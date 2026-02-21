import { useState } from "react";
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unuathorized from "@industry-tool/components/unauthorized";
import PlanetOverview from "@industry-tool/components/pi/PlanetOverview";
import ProfitTable from "@industry-tool/components/pi/ProfitTable";
import SupplyChain from "@industry-tool/components/pi/SupplyChain";
import Navbar from "@industry-tool/components/Navbar";
import Container from "@mui/material/Container";
import Typography from "@mui/material/Typography";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import Box from "@mui/material/Box";

export default function PlanetaryIndustry() {
  const { status } = useSession();
  const [tab, setTab] = useState(() => {
    if (typeof window !== "undefined") {
      return parseInt(localStorage.getItem("pi-tab") || "0", 10);
    }
    return 0;
  });

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unuathorized />;
  }

  const handleTabChange = (_: React.SyntheticEvent, newValue: number) => {
    setTab(newValue);
    localStorage.setItem("pi-tab", String(newValue));
  };

  return (
    <>
      <Navbar />
      <Container maxWidth="xl" sx={{ mt: 2, mb: 4 }}>
        <Typography variant="h5" sx={{ color: "#e2e8f0", mb: 2, fontWeight: 600 }}>
          Planetary Industry
        </Typography>
        <Box sx={{ borderBottom: 1, borderColor: "rgba(148, 163, 184, 0.15)", mb: 2 }}>
          <Tabs
            value={tab}
            onChange={handleTabChange}
            sx={{
              "& .MuiTab-root": { color: "#64748b", textTransform: "none", fontWeight: 500 },
              "& .Mui-selected": { color: "#3b82f6" },
              "& .MuiTabs-indicator": { backgroundColor: "#3b82f6" },
            }}
          >
            <Tab label="Overview" />
            <Tab label="Profit" />
            <Tab label="Supply Chain" />
          </Tabs>
        </Box>
        {tab === 0 && <PlanetOverview embedded />}
        {tab === 1 && <ProfitTable />}
        {tab === 2 && <SupplyChain />}
      </Container>
    </>
  );
}
