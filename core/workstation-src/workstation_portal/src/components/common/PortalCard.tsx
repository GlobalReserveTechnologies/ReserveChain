import React from "react";

interface PortalCardProps {
  title: string;
  subtitle: string;
  tags?: string[];
  ctaLabel: string;
  onClick: () => void;
}

const PortalCard: React.FC<PortalCardProps> = ({
  title,
  subtitle,
  tags = [],
  ctaLabel,
  onClick,
}) => {
  return (
    <div className="portal-card" onClick={onClick}>
      <div className="portal-card__title">{title}</div>
      <div className="portal-card__subtitle">{subtitle}</div>
      {tags.length > 0 && (
        <div className="portal-card__pill-row">
          {tags.map((t) => (
            <div key={t} className="portal-card__pill">
              {t}
            </div>
          ))}
        </div>
      )}
      <div className="portal-card__footer">
        <span className="portal-card__cta">
          {ctaLabel} <span>↗</span>
        </span>
        <span>Secure surface</span>
      </div>
    </div>
  );
};

export default PortalCard;
