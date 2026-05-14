'use client';

interface BearOffZoneProps {
  color: 'white' | 'black';
  count: number;
  x: number;
  y: number;
  isTarget?: boolean;
  onClick?: () => void;
}

export default function BearOffZone({ color, count, x, y, isTarget = false, onClick }: BearOffZoneProps) {
  const fill = color === 'white' ? '#f0f0f0' : '#3a3a3a';
  const text = color === 'white' ? '#3a3a3a' : '#f0f0f0';
  return (
    <g
      data-testid={`bear-off-${color}`}
      onClick={onClick}
      className={onClick ? 'cursor-pointer' : undefined}
    >
      <rect x={x} y={y} width={40} height={80} rx={6} fill={fill} opacity={0.3}
        stroke="#a3b1c6" strokeWidth={1} />
      {isTarget && (
        <rect
          data-testid="valid-bear-off-target"
          x={x - 4}
          y={y - 4}
          width={48}
          height={88}
          rx={8}
          fill="none"
          stroke="#facc15"
          strokeWidth={3}
        />
      )}
      <text x={x + 20} y={y + 50} textAnchor="middle" fontSize={20}
        fontWeight="bold" fill={text}>
        {count}
      </text>
    </g>
  );
}
