import type { FC } from "react"
import { Container, Nav, Navbar } from "react-bootstrap"

export const Header: FC = () => <header style={{ borderBottom: "5px solid #5BA1D4", marginBottom: 24 }}>
    <Navbar expand="lg">
        <Container>
            <Navbar.Brand style={{ fontSize: 36, lineHeight: "48px", fontWeight: "bold", color: "#5B5B5B" }}>VETRYAKI</Navbar.Brand>
            <Navbar.Toggle aria-controls="navbar" />
            <Navbar.Collapse id="navbar" style={{ justifyContent: "end" }}>
                <Nav style={{ alignItems: "center", gap: 20 }}>
                    <Nav.Link href={"/turbines"}>Ветрогенераторы</Nav.Link>
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