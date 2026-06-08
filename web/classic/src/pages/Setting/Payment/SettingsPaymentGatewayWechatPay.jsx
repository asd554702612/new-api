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

import React, { useEffect, useRef, useState } from 'react';
import { Banner, Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import { BookOpen } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
} from '../../../helpers';

export default function SettingsPaymentGatewayWechatPay(props) {
  const { t } = useTranslation();
  const sectionTitle = props.hideSectionTitle ? undefined : t('微信支付设置');
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    WechatPayEnabled: false,
    WechatPayUnitPrice: 0,
    WechatPayAppID: '',
    WechatPayMchID: '',
    WechatPayAPIv3Key: '',
    WechatPayPrivateKey: '',
    WechatPayMerchantSerialNo: '',
    WechatPayPublicKeyID: '',
    WechatPayPublicKey: '',
    WechatPayNotifyURL: '',
    WechatPayReturnURL: '',
    WechatPayNativeEnabled: true,
    WechatPayH5Enabled: true,
    WechatPayJSAPIEnabled: false,
  });
  const formApiRef = useRef(null);

  useEffect(() => {
    if (!props.options || !formApiRef.current) return;
    const currentInputs = {
      WechatPayEnabled: !!props.options.WechatPayEnabled,
      WechatPayUnitPrice:
        props.options.WechatPayUnitPrice !== undefined
          ? parseFloat(props.options.WechatPayUnitPrice)
          : 0,
      WechatPayAppID: props.options.WechatPayAppID || '',
      WechatPayMchID: props.options.WechatPayMchID || '',
      WechatPayAPIv3Key: '',
      WechatPayPrivateKey: '',
      WechatPayMerchantSerialNo:
        props.options.WechatPayMerchantSerialNo || '',
      WechatPayPublicKeyID: props.options.WechatPayPublicKeyID || '',
      WechatPayPublicKey: '',
      WechatPayNotifyURL: props.options.WechatPayNotifyURL || '',
      WechatPayReturnURL: props.options.WechatPayReturnURL || '',
      WechatPayNativeEnabled:
        props.options.WechatPayNativeEnabled !== undefined
          ? !!props.options.WechatPayNativeEnabled
          : true,
      WechatPayH5Enabled:
        props.options.WechatPayH5Enabled !== undefined
          ? !!props.options.WechatPayH5Enabled
          : true,
      WechatPayJSAPIEnabled: !!props.options.WechatPayJSAPIEnabled,
    };
    setInputs(currentInputs);
    formApiRef.current.setValues(currentInputs);
  }, [props.options]);

  const updateOptions = async (options) => {
    const results = await Promise.all(
      options.map((opt) =>
        API.put('/api/option/', {
          key: opt.key,
          value: opt.value,
        }),
      ),
    );
    const errors = results.filter((res) => !res.data.success);
    if (errors.length > 0) {
      errors.forEach((res) => showError(res.data.message));
      return false;
    }
    return true;
  };

  const submitWechatPaySetting = async () => {
    setLoading(true);
    try {
      const options = [
        {
          key: 'WechatPayEnabled',
          value: inputs.WechatPayEnabled ? 'true' : 'false',
        },
        {
          key: 'WechatPayUnitPrice',
          value:
            inputs.WechatPayUnitPrice !== undefined &&
            inputs.WechatPayUnitPrice !== null
              ? inputs.WechatPayUnitPrice.toString()
              : '0',
        },
        { key: 'WechatPayAppID', value: inputs.WechatPayAppID || '' },
        { key: 'WechatPayMchID', value: inputs.WechatPayMchID || '' },
        {
          key: 'WechatPayMerchantSerialNo',
          value: inputs.WechatPayMerchantSerialNo || '',
        },
        {
          key: 'WechatPayPublicKeyID',
          value: inputs.WechatPayPublicKeyID || '',
        },
        {
          key: 'WechatPayNotifyURL',
          value: removeTrailingSlash(inputs.WechatPayNotifyURL || ''),
        },
        {
          key: 'WechatPayReturnURL',
          value: removeTrailingSlash(inputs.WechatPayReturnURL || ''),
        },
        {
          key: 'WechatPayNativeEnabled',
          value: inputs.WechatPayNativeEnabled ? 'true' : 'false',
        },
        {
          key: 'WechatPayH5Enabled',
          value: inputs.WechatPayH5Enabled ? 'true' : 'false',
        },
        {
          key: 'WechatPayJSAPIEnabled',
          value: inputs.WechatPayJSAPIEnabled ? 'true' : 'false',
        },
      ];

      if ((inputs.WechatPayAPIv3Key || '').trim()) {
        options.push({
          key: 'WechatPayAPIv3Key',
          value: inputs.WechatPayAPIv3Key,
        });
      }
      if ((inputs.WechatPayPrivateKey || '').trim()) {
        options.push({
          key: 'WechatPayPrivateKey',
          value: inputs.WechatPayPrivateKey,
        });
      }
      if ((inputs.WechatPayPublicKey || '').trim()) {
        options.push({
          key: 'WechatPayPublicKey',
          value: inputs.WechatPayPublicKey,
        });
      }

      if (await updateOptions(options)) {
        showSuccess(t('更新成功'));
        props.refresh?.();
      }
    } catch (error) {
      showError(t('更新失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={setInputs}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={sectionTitle}>
          <Banner
            type='info'
            icon={<BookOpen size={16} />}
            description={
              <>
                {t('微信支付回调地址')}：
                {props.options.ServerAddress
                  ? removeTrailingSlash(props.options.ServerAddress)
                  : t('网站地址')}
                /api/wechat-pay/notify
              </>
            }
            style={{ marginBottom: 16 }}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch
                field='WechatPayEnabled'
                label={t('启用微信支付')}
                checkedText='｜'
                uncheckedText='〇'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input field='WechatPayAppID' label='AppID' />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input field='WechatPayMchID' label={t('商户号')} />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='WechatPayUnitPrice'
                precision={2}
                label={t('充值价格（x元/美金）')}
                placeholder={t('例如：7，就是7元/美金')}
                extraText={t('按 1 美元对应的站内价格填写')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WechatPayMerchantSerialNo'
                label={t('商户证书序列号')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WechatPayAPIv3Key'
                label='API v3 Key'
                type='password'
                placeholder={t('留空表示保持当前不变')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.TextArea
                field='WechatPayPrivateKey'
                label={t('商户私钥')}
                placeholder={t('留空表示保持当前不变')}
                autosize
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WechatPayPublicKeyID'
                label={t('微信支付公钥 ID')}
                placeholder='PUB_KEY_ID_...'
              />
            </Col>
            <Col xs={24} sm={24} md={16} lg={16} xl={16}>
              <Form.TextArea
                field='WechatPayPublicKey'
                label={t('微信支付公钥')}
                placeholder={t('留空表示保持当前不变')}
                autosize
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='WechatPayNotifyURL'
                label='Notify URL'
                placeholder='https://gateway.example.com/api/wechat-pay/notify'
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='WechatPayReturnURL'
                label='Return URL'
                placeholder='https://gateway.example.com/console/topup'
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch field='WechatPayNativeEnabled' label='Native' />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch field='WechatPayH5Enabled' label='H5' />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch field='WechatPayJSAPIEnabled' label='JSAPI' />
            </Col>
          </Row>
          <Button onClick={submitWechatPaySetting} style={{ marginTop: 16 }}>
            {t('更新微信支付设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
