interface CastMember {
  name: string
  character: string
  order: number
}

interface MovieCastProps {
  cast: CastMember[]
}

export function MovieCast({ cast }: MovieCastProps) {
  if (!cast.length) return null

  return (
    <section className="mt-10">
      <h2 className="text-lg font-semibold mb-3">Casting</h2>
      <div className="flex gap-3 overflow-x-auto pb-2">
        {cast.slice(0, 12).map(member => (
          <div
            key={`${member.name}-${member.order}`}
            className="shrink-0 w-24 text-center"
          >
            <div className="w-16 h-16 rounded-full bg-muted mx-auto flex items-center justify-center overflow-hidden">
              <span className="text-xl font-bold text-muted-foreground">
                {member.name.charAt(0)}
              </span>
            </div>
            <p className="text-xs font-medium mt-2 truncate">{member.name}</p>
            <p className="text-xs text-muted-foreground truncate">{member.character}</p>
          </div>
        ))}
      </div>
    </section>
  )
}
