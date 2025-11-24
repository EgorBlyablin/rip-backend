import { Button, Card } from "react-bootstrap";
import { useHref } from "react-router";
import type { Turbine } from "../api/interfaces";

import defaultImage from "../assets/default.jpg";


export const TurbineCard = (turbine: Turbine) => (
    <Card
        style={{
            maxWidth: 380,
            minWidth: 330,
            backgroundColor: "#F5F7F7",
            border: "none",
        }}
    >
        <Card.Img
            style={{ height: 220, objectFit: "cover" }}
            variant="top"
            src={turbine.image || defaultImage}
            alt={`Фото ${turbine.title}`} />
        <Card.Body className="m-3" style={{ color: "#5B5B5B" }}>
            <Card.Title
                style={{
                    textTransform: "uppercase",
                    fontWeight: "bold",
                }}
                className="mb-4"
            >
                {turbine.title}
            </Card.Title>
            <div className="mb-4">
                <div className="d-flex align-items-center mb-2">
                    <span className="me-2 fw-bold">Мощность:</span>
                    <span className="flex-grow-1 border-bottom border-dotted"></span>
                    <span className="ms-2">
                        {(turbine.power / 1000).toPrecision(2)} кВт
                    </span>
                </div>
                <div className="d-flex align-items-center">
                    <span className="me-2 fw-bold">Мачта:</span>
                    <span className="flex-grow-1 border-bottom border-dotted"></span>
                    <span className="ms-2">{turbine.height} м</span>
                </div>
            </div>
            <Button
                href={useHref(`/turbines/${turbine.id}`)}
                style={{
                    padding: "12px 30px",
                    color: "white",
                    backgroundColor: "#5BA1D4",
                    textTransform: "uppercase",
                    border: "none",
                }}
            >
                Подробнее
            </Button>
        </Card.Body>
    </Card>
);