/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatNumber, formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useUserLeaderboard } from '../hooks/use-user-leaderboard'
import { formatTokens } from '../lib/format'
import type { UserLeaderboardPeriod } from '../types'

const USER_PERIODS: { id: UserLeaderboardPeriod; labelKey: string }[] = [
  { id: 'today', labelKey: 'Today' },
  { id: 'yesterday', labelKey: 'Yesterday' },
]

type UserLeaderboardSectionProps = {
  period: UserLeaderboardPeriod
  onPeriodChange: (period: UserLeaderboardPeriod) => void
}

export function UserLeaderboardSection(props: UserLeaderboardSectionProps) {
  const { t } = useTranslation()
  const query = useUserLeaderboard(props.period)
  const snapshot = query.data?.data

  if (query.isLoading) {
    return (
      <div className='space-y-4'>
        <Skeleton className='h-24 w-full rounded-xl' />
        <Skeleton className='h-[420px] w-full rounded-xl' />
      </div>
    )
  }

  if (!snapshot) {
    return (
      <Card>
        <CardContent className='py-10 text-center'>
          <h2 className='text-foreground text-base font-semibold'>
            {t('Unable to load user leaderboard')}
          </h2>
          <p className='text-muted-foreground mx-auto mt-2 max-w-md text-sm'>
            {query.error instanceof Error
              ? query.error.message
              : t('Unable to load user leaderboard data')}
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <section className='space-y-4'>
      <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
        <div className='border-border/60 inline-flex w-fit rounded-lg border p-1'>
          {USER_PERIODS.map((period) => {
            const isActive = props.period === period.id
            return (
              <button
                key={period.id}
                type='button'
                onClick={() => props.onPeriodChange(period.id)}
                className={cn(
                  'focus-visible:ring-ring/40 rounded-md px-3 py-1.5 text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:outline-none',
                  isActive
                    ? 'bg-foreground text-background'
                    : 'text-muted-foreground hover:text-foreground'
                )}
              >
                {t(period.labelKey)}
              </button>
            )
          })}
        </div>

        <Button
          variant='outline'
          size='sm'
          onClick={() => query.refetch()}
          disabled={query.isFetching}
        >
          <RefreshCw
            className={cn('size-4', query.isFetching && 'animate-spin')}
          />
          {t('Refresh')}
        </Button>
      </div>

      <div className='grid gap-3 md:grid-cols-3'>
        <MetricCard
          title={t('Total tokens')}
          value={formatTokens(snapshot.total_tokens)}
        />
        <MetricCard
          title={t('Total requests')}
          value={formatNumber(snapshot.total_requests)}
        />
        <MetricCard
          title={t('Total consumed quota')}
          value={formatQuota(snapshot.total_quota)}
        />
      </div>

      <Card>
        <CardHeader className='border-b'>
          <CardTitle>{t('User consumption leaderboard')}</CardTitle>
          <CardAction className='text-muted-foreground text-xs tabular-nums'>
            {snapshot.start_date}
          </CardAction>
        </CardHeader>
        <CardContent className='p-0'>
          {snapshot.ranking.length === 0 ? (
            <div className='text-muted-foreground py-14 text-center text-sm'>
              {t('No leaderboard data')}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className='w-16 px-4'>{t('Rank')}</TableHead>
                  <TableHead>{t('User')}</TableHead>
                  <TableHead className='text-right'>{t('Tokens')}</TableHead>
                  <TableHead className='text-right'>{t('Requests')}</TableHead>
                  <TableHead className='px-4 text-right'>
                    {t('Consumed quota')}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {snapshot.ranking.map((row) => (
                  <TableRow key={row.user_id}>
                    <TableCell className='text-muted-foreground px-4 font-mono text-xs'>
                      #{row.rank}
                    </TableCell>
                    <TableCell className='font-medium'>
                      {row.display_name}
                    </TableCell>
                    <TableCell className='text-right font-mono'>
                      {formatTokens(row.tokens)}
                    </TableCell>
                    <TableCell className='text-right font-mono'>
                      {formatNumber(row.requests)}
                    </TableCell>
                    <TableCell className='px-4 text-right font-mono'>
                      {formatQuota(row.quota)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </section>
  )
}

function MetricCard(props: { title: string; value: string }) {
  return (
    <Card size='sm'>
      <CardHeader>
        <CardTitle className='text-muted-foreground text-xs font-medium uppercase'>
          {props.title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className='text-foreground font-mono text-xl font-semibold tabular-nums'>
          {props.value}
        </div>
      </CardContent>
    </Card>
  )
}
