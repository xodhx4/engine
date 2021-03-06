/*
 * Copyright 2018 It-chain
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api_test

import (
	"errors"
	"fmt"
	"testing"

	"os"

	"encoding/hex"

	"github.com/it-chain/engine/common"
	"github.com/it-chain/engine/ivm"
	"github.com/it-chain/engine/ivm/api"
	"github.com/it-chain/engine/ivm/infra/git"
	"github.com/it-chain/engine/ivm/infra/tesseract"
	"github.com/stretchr/testify/assert"
)

func TestICodeApi_Deploy(t *testing.T) {

	savePath := os.Getenv("GOPATH") + "/src/github.com/it-chain/engine/.tmp/"
	defer os.RemoveAll(savePath)

	sshPath := "./id_rsa"
	err, tearDown1 := generatePriKey(sshPath)
	assert.NoError(t, err)
	defer tearDown1()

	api, containerService := setUp(t)

	icode, err := api.Deploy(savePath, "github.com/junbeomlee/learn-icode", sshPath, "")
	defer api.UnDeploy(icode.ID)

	assert.NoError(t, err)
	assert.Equal(t, icode.RepositoryName, "learn-icode")
	assert.Equal(t, icode.GitUrl, "github.com/junbeomlee/learn-icode")
	assert.Equal(t, containerService.GetRunningICodeList()[0].ID, icode.ID)
}

func TestICodeApi_UnDeploy(t *testing.T) {
	savePath := os.Getenv("GOPATH") + "/src/github.com/it-chain/engine/.tmp/"
	defer os.RemoveAll(savePath)

	sshPath := "./id_rsa"
	err, tearDown1 := generatePriKey(sshPath)
	assert.NoError(t, err)
	defer tearDown1()

	api, containerService := setUp(t)
	icode, err := api.Deploy(savePath, "github.com/junbeomlee/learn-icode", sshPath, "")
	assert.NoError(t, err)
	assert.Equal(t, containerService.GetRunningICodeList()[0].ID, icode.ID)

	err = api.UnDeploy(icode.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(containerService.GetRunningICodeList()))
}

func TestICodeApi_ExecuteRequest(t *testing.T) {
	savePath := os.Getenv("GOPATH") + "/src/github.com/it-chain/engine/.tmp/"
	defer os.RemoveAll(savePath)

	sshPath := "./id_rsa"
	err, tearDown1 := generatePriKey(sshPath)
	assert.NoError(t, err)
	defer tearDown1()

	api, _ := setUp(t)
	icode, err := api.Deploy(savePath, "github.com/junbeomlee/learn-icode", sshPath, "")
	defer api.UnDeploy(icode.ID)

	result, err := api.ExecuteRequest(ivm.Request{
		ICodeID:  icode.ID,
		Function: "initA",
		Type:     "invoke",
		Args:     []string{},
	})

	assert.NoError(t, err)
	assert.Equal(t, result.Err, "")
}

func TestICodeApi_ExecuteRequestList(t *testing.T) {
	savePath := os.Getenv("GOPATH") + "/src/github.com/it-chain/engine/.tmp/"
	defer os.RemoveAll(savePath)

	sshPath := "./id_rsa"
	err, tearDown1 := generatePriKey(sshPath)
	assert.NoError(t, err)
	defer tearDown1()

	api, _ := setUp(t)
	icode, err := api.Deploy(savePath, "github.com/junbeomlee/learn-icode", sshPath, "")
	defer api.UnDeploy(icode.ID)

	results := api.ExecuteRequestList([]ivm.Request{
		ivm.Request{
			ICodeID:  icode.ID,
			Function: "initA",
			Type:     "invoke",
			Args:     []string{},
		},
		ivm.Request{
			ICodeID:  icode.ID,
			Function: "incA",
			Type:     "invoke",
			Args:     []string{},
		},
	})

	assert.Equal(t, len(results), 2)
	for _, result := range results {
		assert.Equal(t, result.Err, "")
	}
}

func TestICodeApi_GetRunningICodeIDList(t *testing.T) {
	savePath := os.Getenv("GOPATH") + "/src/github.com/it-chain/engine/.tmp/"
	defer os.RemoveAll(savePath)

	sshPath := "./id_rsa"
	err, tearDown1 := generatePriKey(sshPath)
	assert.NoError(t, err)
	defer tearDown1()

	api, containerService := setUp(t)
	icode, err := api.Deploy(savePath, "github.com/junbeomlee/learn-icode", sshPath, "")
	defer api.UnDeploy(icode.ID)
	assert.NoError(t, err)
	assert.Equal(t, containerService.GetRunningICodeList()[0].ID, icode.ID)

	icodeIDs := api.GetRunningICodeList()
	assert.NotNil(t, icodeIDs)
	assert.Equal(t, icodeIDs[0].ID, icode.ID)
}

func setUp(t *testing.T) (*api.ICodeApi, *tesseract.ContainerService) {
	GOPATH := os.Getenv("GOPATH")

	if GOPATH == "" {
		t.Fatal(errors.New("need go path"))
	}

	// git generate
	storeApi := git.NewRepositoryService()
	containerService := tesseract.NewContainerService()
	eventService := common.NewEventService("", "Event")
	icodeApi := api.NewICodeApi(containerService, storeApi, eventService)

	return &icodeApi, containerService
}

func generatePriKey(path string) (error, func() error) {
	_, err := os.Stat(path)
	if err == nil {
		os.Remove(path)
	}

	file, err := os.Create(path)
	if err != nil {
		return err, nil
	}

	src := "2d2d2d2d2d424547494e205253412050524956415445204b45592d2d2d2d2d0a4d4949456f67494241414b4341514541784237474d566a7135376e314f49646f736950354341677141506b53334641417231416b4c4f6e48424331484d636c530a66586e7732745a54722b41386474764e37442b71364c5a793666634551623555546271706e4a51706a39304e356268366f4f317a7538433649586c52363237330a586f525536334f784170766f5a772f4d5a337a2f3467533679324343466c7a59426c743635336663787a754e7a3359374275325a476f5334732f4135783958520a656962622b6c74596732655658514562627936794d74455862557850652f7757664c583268544630316645374a67517a35766f526644356e3164396c326f37490a53734137533135464535334254673954334f6f553161553964586368316f32786e6d5142647a5465534a466d54766a6f624f69426650315665444d6d6a71624e0a4c335174617a3352574a4266756d6e45647969737870746d55346d6142624a7a536e2b634b51494441514142416f4942414632436a7431596d436945386664530a47516c58505a596d7a6d4249596b584a6e346e336e45674e37325a2b63454f38796967707a44324c6b3774344831784d305a4b6a694d6f4d744233364f5831660a55724c394859496134765a4659437234477742414e373539316b472f70742b717553664830505779342b4e716b7855513431553074497a2f3146444559304a6d0a596c6f6c7043525a636c744d65674642546b4f765a69444f78344b44596b4a6c436d6f65616e36325a384d5a775a6d7532597a37756a32455a7666764f586a730a7a3749504b5a6a6f524c2b447a577142466e5347706d6b5876324c7a5334546c363041747545354768494558392f777053346b583568376734314b4d2f3375620a336145764a7949486841545855626373344b436a70634a75757a61463038476d63614b79424439416a4d6b6f595553684945716d2b566d53677533524d5877520a43372f77344145436759454135794167786a744e4a364c5435346a446c5a6c4132544f61735a6d665654686c504c5674503866547847426650493377454954370a4a6c3355526e53392f6e6855684d4c33477465684355786b6f7173304e6256672f6159516b78696156394c65622f357439444e31636e482b47616141697656410a5149334d35304c4a6a344b5251646b664d6e6467614f46475a7150684e3645307439413070764f7a357832456555534958546669597345436759454132546f780a2f5675646f63434558356b385271616c775157486e7946455664564c567076474e7739576f5837394f3847315338356d784870622b4951736c4f31795954692b0a65654b5a4239496e77483532637168453470547637592f6b76413379334947456d4a75345772576f77697768746d5873586c73544b532b7431424b35526364460a352f744b6f693937554f383848527936446b4a3863645268544a7839756b57716579437132326b43675941736c58503943543975332b674568387443746c64650a4471684f6a6858414f4b712b7454696e7a77493470575a3570642b6a4d43504b574e737a354230715530667166446c7967686e635531497556747778614257580a6d45736d4e4e3742426a70475845775669542b6b6e66796f4d67676c7866316f396e474b51735869327772754b74587278443969752b4836747134684c7757650a56356c77677834322f4f6972413939534c412b4e67514b4267446d36425637573465554354437337685a45673642754c5a4b63644b4250485175595a4c3275690a58397336374144645556683732554f4e594c4f434c4862485177596a466a7339784830586c416a4c6b703656715069747137547438464d7051636a6e676c30720a784b6f5762477074582b67673364655654466f396d5777714c61496c65715a5457566f5156433046356d7532487074376636616755647353477a644e48436a730a58587442416f474146476f6559706c39562b7a54525974303366582f63504b34546c335a666a48325a4b67704372484951786e6a4957426f37706d31734159590a62764f5743757a704b486d6d4e794443504154707851564d6d417752416762635664736365342f7930624b467161784c387a6c3731614a6e324e3769344567540a6a6262346155396c39534d70314366744933355766394d76714341496b5476372b675351674a726d4161636e696a436f6431343d0a2d2d2d2d2d454e44205253412050524956415445204b45592d2d2d2d2d0a"
	p, _ := hex.DecodeString(src)

	file.WriteString(fmt.Sprintf("%s", p))
	return nil, func() error {

		if err := os.RemoveAll(path); err != nil {
			return err
		}

		if err := file.Close(); err != nil {
			return err
		}

		return nil
	}
}
