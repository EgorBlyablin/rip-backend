import { Container, Row, Col } from "react-bootstrap";
import { Header } from "../components/header";

export const MainPage = () => {
    return (
        <div>
            <div style={{position: "absolute", width: "100%", zIndex: 10}}>
                <Header mode={"main-page"} />
            </div>
            <div className="position-relative min-vh-100 d-flex align-items-center justify-content-center overflow-hidden">
                <video
                    autoPlay
                    muted
                    loop
                    playsInline
                    className="position-absolute top-0 start-0 w-100 h-100"
                    style={{ objectFit: "cover", zIndex: -2 }}
                    poster={`${import.meta.env.BASE_URL}first-frame.webp`} // Optional fallback
                >
                    <source src={`${import.meta.env.BASE_URL}video.webm`} type="video/webm" />
                </video>

                {/* Dark Overlay for Better Text Contrast */}
                <div
                    className="position-absolute top-0 start-0 w-100 h-100"
                    style={{ backgroundColor: "rgba(0, 0, 0, 0.2)", zIndex: -1 }}
                ></div>

                {/* Centered Content */}
                <Container fluid className="text-white text-center p-4">
                    <Row>
                        <Col>
                            <h1 className="display-4 mb-3 fw-bold">Добро пожаловать в VETRYAKI</h1>
                            <p className="lead mb-4 opacity-90">
                                Ваш надежный партнер в мире чистой энергии
                            </p>
                        </Col>
                    </Row>
                </Container>
            </div>
        </div>
    );
};