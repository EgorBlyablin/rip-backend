import type { FC } from "react"
import { Container, Nav, Navbar } from "react-bootstrap"

export const Header: FC<{ mode: "normal" | "main-page" }> = ({ mode = "normal" }) => (
    <header style={{
        borderBottom: mode === "normal" ? "5px solid #5BA1D4" : undefined,
        marginBottom: 24
    }}>
        <Navbar expand="lg">
            <Container>
                <Navbar.Brand style={{
                    fontSize: 36,
                    lineHeight: "48px",
                    fontWeight: "bold",
                    color: mode === "normal" ? undefined : "white",
                }}>VETRYAKI</Navbar.Brand>
                <Navbar.Toggle aria-controls="navbar" />
                <Navbar.Collapse id="navbar" style={{ justifyContent: "end" }}>
                    <Nav style={{ alignItems: "center", gap: 20 }}>
                        <Nav.Link href={"/turbines"} style={{ color: mode === "normal" ? "black" : "white" }}>Ветрогенераторы</Nav.Link>
                        <Nav.Link href={"/"} style={{
                            padding: "12px 30px",
                            color: "white",
                            backgroundColor: "#5BA1D4",
                            borderRadius: 6
                        }}>Домой</Nav.Link>
                    </Nav>
                </Navbar.Collapse>
            </Container>
        </Navbar>
    </header >
)