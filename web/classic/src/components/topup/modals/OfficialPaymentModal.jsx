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
import { Modal, Typography } from '@douyinfe/semi-ui';
import { QRCodeSVG } from 'qrcode.react';

const OfficialPaymentModal = ({ t, visible, onCancel, payment }) => {
  const qrValue = payment?.qrValue || '';

  return (
    <Modal
      title={t('扫码支付')}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      centered
      size='small'
    >
      <div className='flex flex-col items-center gap-3 py-2'>
        {qrValue && (
          <div className='rounded-lg border border-gray-100 bg-white p-4'>
            <QRCodeSVG value={qrValue} size={220} />
          </div>
        )}
        <Typography.Text type='tertiary' size='small' align='center'>
          {t('支付成功后，订单会自动到账。')}
        </Typography.Text>
      </div>
    </Modal>
  );
};

export default OfficialPaymentModal;
