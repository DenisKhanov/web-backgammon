'use client';

interface BearOffZoneProps {
  color: 'white' | 'black';
  count: number;
  x: number;
  y: number;
}

export default function BearOffZone({ color, count, x, y }: BearOffZoneProps) {
  const fill = color === 'white' ? '#f0f0f0' : '#3a3a3a';
  const text = color === 'white' ? '#3a3a3a' : '#f0f0f0';
  return (
    <g>
      <rect x={x} y={y} width={40} height={80} rx={6} fill={fill} opacity={0.3}
        stroke="#a3b1c6" strokeWidth={1} />
      <text x={x + 20} y={y + 50} textAnchor="middle" fontSize={20}
        fontWeight="bold" fill={text}>
        {count}
      </text>
    </g>
  );
}
