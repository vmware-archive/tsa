package tsa_test

import (
	"encoding/json"

	"github.com/concourse/tsa"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/concourse/atc"
	"github.com/concourse/tsa/tsafakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/rata"
)

var _ = Describe("Check Worker", func() {
	var (
		checker *tsa.CheckWorker

		logger             *lagertest.TestLogger
		worker             atc.Worker
		fakeTokenGenerator *tsafakes.FakeTokenGenerator
		fakeATC            *ghttp.Server
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		worker = atc.Worker{
			Name: "some-worker",
		}
		fakeTokenGenerator = new(tsafakes.FakeTokenGenerator)
		fakeTokenGenerator.GenerateSystemTokenReturns("yo", nil)

		fakeATC = ghttp.NewServer()

		atcEndpoint := rata.NewRequestGenerator(fakeATC.URL(), atc.Routes)

		checker = &tsa.CheckWorker{
			ATCEndpoint:    atcEndpoint,
			TokenGenerator: fakeTokenGenerator,
		}
	})

	AfterEach(func() {
		fakeATC.Close()
	})

	It("tells the ATC to list the worker", func() {
		workers := []atc.Worker{}
		workers = append(workers, atc.Worker{Name: "test-worker"})
		data, err := json.Marshal(workers)
		Ω(err).ShouldNot(HaveOccurred())

		fakeATC.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/v1/workers"),
			ghttp.VerifyHeaderKV("Authorization", "Bearer yo"),
			ghttp.RespondWith(200, data, nil),
		))

		err = checker.CheckStatus(logger, worker)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeATC.ReceivedRequests()).To(HaveLen(1))
	})

	Context("when the ATC does not respond to retire the worker", func() {
		BeforeEach(func() {
			fakeATC.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/workers"),
				ghttp.RespondWith(500, nil, nil),
			))
		})

		It("errors", func() {
			err := checker.CheckStatus(logger, worker)
			Expect(err).To(HaveOccurred())

			Expect(err).To(MatchError(ContainSubstring("500")))
			Expect(fakeATC.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
