/*
Copyright (C) 2025 QuantumNous

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

import React from 'react';
import { Input, InputNumber, Modal, Typography } from '@douyinfe/semi-ui';
import { WalletCards } from 'lucide-react';

const AffiliateWithdrawModal = ({
  t,
  visible,
  onOk,
  onCancel,
  userState,
  renderQuota,
  withdrawAmount,
  setWithdrawAmount,
  payoutNote,
  setPayoutNote,
  minQuota,
  helpText,
  loading,
}) => {
  return (
    <Modal
      title={
        <div className='flex items-center'>
          <WalletCards className='mr-2' size={18} />
          {t('邀请返利提现')}
        </div>
      }
      visible={visible}
      onOk={onOk}
      onCancel={onCancel}
      maskClosable={false}
      centered
      confirmLoading={loading}
    >
      <div className='space-y-4'>
        <div>
          <Typography.Text strong className='block mb-2'>
            {t('可用邀请额度')}
          </Typography.Text>
          <Input value={renderQuota(userState?.user?.aff_quota || 0)} disabled />
        </div>
        <div>
          <Typography.Text strong className='block mb-2'>
            {t('提现额度')} · {t('最低') + renderQuota(minQuota || 0)}
          </Typography.Text>
          <InputNumber
            min={minQuota || 0}
            max={userState?.user?.aff_quota || 0}
            value={withdrawAmount}
            onChange={(value) => setWithdrawAmount(value)}
            className='w-full'
          />
        </div>
        <div>
          <Typography.Text strong className='block mb-2'>
            {t('收款说明')}
          </Typography.Text>
          <Input
            value={payoutNote}
            onChange={setPayoutNote}
            placeholder={t('请填写微信收款账号或联系方式')}
          />
          {helpText && (
            <Typography.Text type='tertiary' className='block mt-2 text-xs'>
              {helpText}
            </Typography.Text>
          )}
        </div>
      </div>
    </Modal>
  );
};

export default AffiliateWithdrawModal;
