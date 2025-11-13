import { Container, Row, Col, Button } from "react-bootstrap";

export const MainPage = () => {
    return (
        <div className="position-relative min-vh-100 d-flex align-items-center justify-content-center overflow-hidden">
            <video
                autoPlay
                muted
                loop
                playsInline
                className="position-absolute top-0 start-0 w-100 h-100"
                style={{ objectFit: "cover", zIndex: -2 }}
                poster="/first-frame.webp" // Optional fallback
            >
                <source src="/video.webm" type="video/webm" />
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
                        <Button tabIndex={1} variant="outline-light" size="lg" href="/turbines">
                            НАЧАТЬ
                        </Button>
                    </Col>
                </Row>
            </Container>
        </div>
    );
};