'use client'
import React from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2, Users, Film, Eye, TrendingUp } from 'lucide-react'
import { apiClient } from '@/lib/api'
import {
  AreaChart, Area,
  BarChart, Bar,
  PieChart, Pie, Cell,
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend,
} from 'recharts'

interface AdminStats {
  total_users: number
  total_films: number
  total_watches: number
  films_by_status: { status: string; count: number }[]
  watches_per_day: { day: string; count: number }[]
  top_films: { title: string; watches: number }[]
  registrations_per_month: { month: string; count: number }[]
}

const STATUS_COLORS: Record<string, string> = {
  ready:       'oklch(0.627 0.194 149.214)',
  downloading: 'oklch(0.488 0.243 264.376)',
  pending:     'oklch(0.6 0.05 264)',
  error:       'oklch(0.577 0.245 27.325)',
}

function StatCard({ icon: Icon, label, value }: { icon: React.ElementType; label: string; value: number }) {
  return (
    <Card>
      <CardContent className="flex items-center gap-4 pt-6">
        <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-sidebar-primary/10">
          <Icon className="size-6 text-sidebar-primary" />
        </div>
        <div>
          <p className="text-sm text-muted-foreground">{label}</p>
          <p className="text-2xl font-bold tabular-nums">{value.toLocaleString()}</p>
        </div>
      </CardContent>
    </Card>
  )
}

export default function AdminOverviewPage() {
  const { t } = useTranslation()
  const [stats, setStats] = React.useState<AdminStats | null>(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    apiClient.get<{ data: AdminStats }>('/admin/stats')
      .then((json) => setStats(json.data ?? null))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="flex justify-center py-24">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!stats) {
    return <p className="text-muted-foreground text-sm py-8 text-center">{t('admin.stats_error')}</p>
  }

  const readyCount = stats.films_by_status.find((s) => s.status === 'ready')?.count ?? 0

  return (
    <div className="space-y-6">
      {/* KPI cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard icon={Users}     label={t('admin.stat_users')}    value={stats.total_users} />
        <StatCard icon={Film}      label={t('admin.stat_films')}    value={stats.total_films} />
        <StatCard icon={TrendingUp} label={t('admin.stat_ready')}   value={readyCount} />
        <StatCard icon={Eye}       label={t('admin.stat_watches')}  value={stats.total_watches} />
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        {/* Activity — area chart */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">{t('admin.chart_activity')}</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={220}>
              <AreaChart data={stats.watches_per_day} margin={{ top: 4, right: 8, left: -24, bottom: 0 }}>
                <defs>
                  <linearGradient id="watchGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%"  stopColor="oklch(0.488 0.243 264.376)" stopOpacity={0.25} />
                    <stop offset="95%" stopColor="oklch(0.488 0.243 264.376)" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                <XAxis
                  dataKey="day"
                  tick={{ fontSize: 11 }}
                  tickFormatter={(v: string) => v.slice(5)}
                  className="text-muted-foreground"
                />
                <YAxis tick={{ fontSize: 11 }} allowDecimals={false} className="text-muted-foreground" />
                <Tooltip
                  contentStyle={{ fontSize: 12 }}
                  labelFormatter={(v: string) => v}
                  formatter={(v: number) => [v, t('admin.stat_watches')]}
                />
                <Area
                  type="monotone"
                  dataKey="count"
                  stroke="oklch(0.488 0.243 264.376)"
                  strokeWidth={2}
                  fill="url(#watchGrad)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* Films by status — pie chart */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t('admin.chart_status')}</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={220}>
              <PieChart>
                <Pie
                  data={stats.films_by_status}
                  dataKey="count"
                  nameKey="status"
                  cx="50%"
                  cy="50%"
                  innerRadius={55}
                  outerRadius={80}
                  paddingAngle={3}
                >
                  {stats.films_by_status.map((entry) => (
                    <Cell
                      key={entry.status}
                      fill={STATUS_COLORS[entry.status] ?? 'oklch(0.6 0.05 264)'}
                    />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{ fontSize: 12 }}
                  formatter={(v: number, name: string) => [v, t(`admin.status_${name}`) || name]}
                />
                <Legend
                  iconType="circle"
                  iconSize={8}
                  formatter={(value: string) => t(`admin.status_${value}`) || value}
                  wrapperStyle={{ fontSize: 12 }}
                />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        {/* Top films — horizontal bar */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t('admin.chart_top_films')}</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={220}>
              <BarChart
                data={stats.top_films}
                layout="vertical"
                margin={{ top: 4, right: 16, left: 8, bottom: 0 }}
              >
                <CartesianGrid strokeDasharray="3 3" horizontal={false} className="stroke-border" />
                <XAxis type="number" tick={{ fontSize: 11 }} allowDecimals={false} className="text-muted-foreground" />
                <YAxis
                  type="category"
                  dataKey="title"
                  width={120}
                  tick={{ fontSize: 11 }}
                  tickFormatter={(v: string) => v.length > 16 ? v.slice(0, 15) + '…' : v}
                  className="text-muted-foreground"
                />
                <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number) => [v, t('admin.stat_watches')]} />
                <Bar dataKey="watches" fill="oklch(0.488 0.243 264.376)" radius={[0, 4, 4, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* Registrations per month — bar chart */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t('admin.chart_registrations')}</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={220}>
              <BarChart
                data={stats.registrations_per_month}
                margin={{ top: 4, right: 8, left: -24, bottom: 0 }}
              >
                <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                <XAxis dataKey="month" tick={{ fontSize: 11 }} className="text-muted-foreground" />
                <YAxis tick={{ fontSize: 11 }} allowDecimals={false} className="text-muted-foreground" />
                <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number) => [v, t('admin.stat_users')]} />
                <Bar dataKey="count" fill="oklch(0.577 0.245 27.325)" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
