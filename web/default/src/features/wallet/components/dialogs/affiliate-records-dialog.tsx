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
import { useCallback, useEffect, useState } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { StatusBadge } from '@/components/status-badge'
import { getAffiliateRecords } from '../../api'
import { formatTimestamp } from '../../lib/billing'
import type { AffiliateLedgerRecord } from '../../types'

interface AffiliateRecordsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const PAGE_SIZE = 10

function getActionLabel(action: AffiliateLedgerRecord['action']) {
  if (action === 'accrue') return 'Cashback'
  if (action === 'transfer') return 'Transfer'
  return action
}

export function AffiliateRecordsDialog({
  open,
  onOpenChange,
}: AffiliateRecordsDialogProps) {
  const { t } = useTranslation()
  const [records, setRecords] = useState<AffiliateLedgerRecord[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const fetchRecords = useCallback(async () => {
    if (!open) return
    try {
      setLoading(true)
      const response = await getAffiliateRecords(page, PAGE_SIZE)
      if (response.success && response.data) {
        setRecords(response.data.items ?? [])
        setTotal(response.data.total ?? 0)
      }
    } finally {
      setLoading(false)
    }
  }, [open, page])

  useEffect(() => {
    fetchRecords()
  }, [fetchRecords])

  useEffect(() => {
    if (open) setPage(1)
  }, [open])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='flex max-h-[calc(100dvh-2rem)] flex-col max-sm:h-dvh max-sm:w-screen max-sm:max-w-none max-sm:rounded-none max-sm:p-4 sm:max-w-3xl'>
        <DialogHeader>
          <DialogTitle>{t('Referral Records')}</DialogTitle>
          <DialogDescription>
            {t('Review cashback accruals and reward transfers')}
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className='h-[calc(100dvh-13rem)] pr-3 sm:h-[440px] sm:pr-4'>
          {loading ? (
            <div className='space-y-3'>
              {Array.from({ length: 5 }).map((_, index) => (
                <div key={index} className='rounded-lg border p-3'>
                  <Skeleton className='h-4 w-40' />
                  <div className='mt-3 grid grid-cols-2 gap-3 sm:grid-cols-3'>
                    <Skeleton className='h-3 w-full' />
                    <Skeleton className='h-3 w-full' />
                    <Skeleton className='h-3 w-full' />
                  </div>
                </div>
              ))}
            </div>
          ) : records.length === 0 ? (
            <div className='text-muted-foreground flex h-[280px] flex-col items-center justify-center text-center'>
              <p className='text-sm font-medium'>
                {t('No referral records found')}
              </p>
              <p className='mt-1 text-xs'>
                {t('Referral cashback and transfers will appear here')}
              </p>
            </div>
          ) : (
            <div className='space-y-3'>
              {records.map((record) => (
                <div
                  key={record.id}
                  className='hover:bg-muted/50 rounded-lg border p-3 transition-colors sm:p-4'
                >
                  <div className='flex items-start justify-between gap-3'>
                    <div className='min-w-0'>
                      <div className='flex items-center gap-2'>
                        <StatusBadge
                          label={t(getActionLabel(record.action))}
                          variant={
                            record.action === 'accrue' ? 'success' : 'neutral'
                          }
                          showDot
                          copyable={false}
                        />
                        {record.source_order_trade_no ? (
                          <code className='text-muted-foreground truncate font-mono text-xs'>
                            {record.source_order_trade_no}
                          </code>
                        ) : null}
                      </div>
                      <div className='text-muted-foreground mt-1 text-xs'>
                        {formatTimestamp(record.created_at)}
                      </div>
                    </div>
                    <div className='shrink-0 text-right text-sm font-semibold'>
                      {formatQuota(record.quota)}
                    </div>
                  </div>

                  <div className='mt-3 grid grid-cols-2 gap-3 sm:grid-cols-3'>
                    <div className='space-y-1'>
                      <Label className='text-muted-foreground text-xs'>
                        {t('Balance After')}
                      </Label>
                      <div className='text-sm font-medium'>
                        {formatQuota(record.balance_after)}
                      </div>
                    </div>
                    <div className='space-y-1'>
                      <Label className='text-muted-foreground text-xs'>
                        {t('Total Earned')}
                      </Label>
                      <div className='text-sm font-medium'>
                        {formatQuota(record.history_after)}
                      </div>
                    </div>
                    <div className='space-y-1'>
                      <Label className='text-muted-foreground text-xs'>
                        {t('Payment Method')}
                      </Label>
                      <div className='truncate text-sm font-medium'>
                        {record.payment_method || '-'}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>

        <div className='flex items-center justify-between border-t pt-3'>
          <div className='text-muted-foreground text-xs'>
            {t('{{page}} / {{totalPages}}', { page, totalPages })}
          </div>
          <div className='flex items-center gap-2'>
            <Button
              variant='outline'
              size='sm'
              onClick={() => setPage((current) => Math.max(1, current - 1))}
              disabled={page <= 1 || loading}
            >
              <ChevronLeft className='size-4' />
            </Button>
            <Button
              variant='outline'
              size='sm'
              onClick={() =>
                setPage((current) => Math.min(totalPages, current + 1))
              }
              disabled={page >= totalPages || loading}
            >
              <ChevronRight className='size-4' />
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
